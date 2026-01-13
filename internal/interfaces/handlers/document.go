package handlers

import (
	"bytes"
	"context"
	"crypto/md5" // __ADDED__
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/usecase/documents_uc"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

// Storage abstrato compatível com s3storage.Storage
type ObjectStorage interface {
	Save(ctx context.Context, key string, r io.Reader) (publicURL string, err error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	URL(key string) string
}

type DocumentHandler struct {
	CreateUC        *documents_uc.CreateDocumentUseCase
	UpdateUC        *documents_uc.UpdateDocumentUseCase
	DeleteUC        *documents_uc.DeleteDocumentUseCase
	ListUC          *documents_uc.ListDocumentsUseCase
	FindUC          *documents_uc.FindDocumentByIDUseCase
	LockConcurrency *utils.KeyedLocker
	Storage         ObjectStorage
	MaxUploadSize   int64
	AllowedExts     map[string]bool
	AllowedTypes    map[string]bool

	TrashPrefix string // __ADDED__: prefixo da “lixeira” (backup temporário no bucket)
}

func NewDocumentHandler(
	createUC *documents_uc.CreateDocumentUseCase,
	updateUC *documents_uc.UpdateDocumentUseCase,
	deleteUC *documents_uc.DeleteDocumentUseCase,
	findOneUC *documents_uc.FindDocumentByIDUseCase,
	findAllUC *documents_uc.ListDocumentsUseCase,
	lock *utils.KeyedLocker,
	storage ObjectStorage,
) *DocumentHandler {
	return &DocumentHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		FindUC:          findOneUC,
		ListUC:          findAllUC,
		LockConcurrency: lock,
		Storage:         storage,
		MaxUploadSize:   2 * 1024 * 1024, // 2MB
		AllowedExts: map[string]bool{
			".pdf": true, ".jpg": true, ".jpeg": true, ".png": true,
		},
		AllowedTypes: map[string]bool{
			"application/pdf": true, "image/jpeg": true, "image/png": true,
		},
		TrashPrefix: "trash", // __ADDED__
	}
}

// --------------------------------- CONSISTENCY HELPERS (S3 <-> DB) ---------------------------------

// Detecta erro de objeto inexistente no S3/MinIO de forma tolerante
func isNoSuchKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nosuchkey") || strings.Contains(msg, "statuscode: 404")
}

// backupObject: agora verifica existência antes e ignora “NoSuchKey”
func (h *DocumentHandler) backupObject(ctx context.Context, key string) (trashKey string, restore func(context.Context) error, cleanup func(context.Context) error, err error) {
	// Verifica se o objeto existe; se não existir, não é erro — apenas não há backup
	exists, exErr := h.Storage.Exists(ctx, key)
	if exErr != nil {
		// erro de metadados; devolver para o chamador decidir
		return "", nil, nil, fmt.Errorf("exists: %w", exErr)
	}
	if !exists {
		// Sem objeto: devolve no-ops de restore/cleanup
		restore = func(context.Context) error { return nil }
		cleanup = func(context.Context) error { return nil }
		return "", restore, cleanup, nil
	}

	// Há objeto: tentar abrir e copiar
	rc, err := h.Storage.Open(ctx, key)
	if err != nil {
		if isNoSuchKey(err) {
			// Corrida: objeto sumiu entre Exists e Open — trata como “sem backup”
			restore = func(context.Context) error { return nil }
			cleanup = func(context.Context) error { return nil }
			return "", restore, cleanup, nil
		}
		return "", nil, nil, fmt.Errorf("abrir original: %w", err)
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", nil, nil, fmt.Errorf("copiar original: %w", err)
	}

	base := filepath.Base(key)
	trashKey = fmt.Sprintf("%s/%s_%s", strings.TrimSuffix(h.TrashPrefix, "/"), uuid.NewString(), base)

	if _, err := h.Storage.Save(ctx, trashKey, bytes.NewReader(buf.Bytes())); err != nil {
		return "", nil, nil, fmt.Errorf("gravar backup: %w", err)
	}

	restore = func(ctx context.Context) error {
		_, err := h.Storage.Save(ctx, key, bytes.NewReader(buf.Bytes()))
		return err
	}
	cleanup = func(ctx context.Context) error {
		return h.Storage.Delete(ctx, trashKey)
	}
	return trashKey, restore, cleanup, nil
}

// __ADDED__: calcula MD5 e tamanho, devolve um ReadSeeker reposicionado para upload/uso posterior.
func computeMD5AndSize(r io.Reader) (sum string, size int64, reread io.ReadSeeker, err error) {
	var mem bytes.Buffer
	n, err := io.Copy(&mem, r)
	if err != nil {
		return "", 0, nil, err
	}
	h := md5.New()
	if _, err := h.Write(mem.Bytes()); err != nil {
		return "", 0, nil, err
	}
	sum = fmt.Sprintf("%x", h.Sum(nil))
	size = n
	reread = bytes.NewReader(mem.Bytes())
	return
}

// --------------------------------- READ ---------------------------------

// GetDocumentByID godoc
// @Summary Busca um documento pelo ID
// @Tags Documentos
// @Produce json
// @Param id path string true "ID do documento"
// @Success 200 {object} dtos.DocumentResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /documents/{id} [get]
func (h *DocumentHandler) FindByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	doc, err := h.FindUC.Execute(ctx, id)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusNotFound, utils.Payload{Error: "not_found", Message: "Documento não encontrado"})
		return
	}
	c.JSON(http.StatusOK, dtos.ToDocumentResponseDTO(doc))
}

// ListDocumentsByOwner godoc
// @Summary Lista documentos de um proprietário (com paginação)
// @Description Retorna os documentos pertencentes ao **owner_id** informado, ordenados por *type ASC* e *uploaded_at DESC*.
// @Tags Documentos
// @Produce json
// @Param owner_id query string true "UUID do proprietário (owner)"
// @Param limit query int false "Limite de resultados (ex.: 20)"
// @Param offset query int false "Offset para paginação (ex.: 0)"
// @Success 200 {array} dtos.DocumentResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /documents [get]
func (h *DocumentHandler) ListByOwnerID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.ListDocumentsByOwnerDTO
	if err := c.ShouldBindQuery(&dto); err != nil {
		if fields := utils.HumanizeValidation(dto, err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação dos campos",
				Message: "Dados inválidos. Corrija os campos destacados.",
				Fields:  fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	ownerID, err := uuid.Parse(dto.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do proprietário deve ser um UUID válido.", Value: dto.OwnerID},
			},
		})
		return
	}

	pg := utils.PaginationInput(c)
	list, err := h.ListUC.Execute(ctx, ownerID, pg.Limit, pg.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos documentos.",
			Fields:  []utils.FieldError{},
		})
		return
	}
	c.JSON(http.StatusOK, dtos.ToDocumentResponseDTOList(list))
}

// --------------------------------- DELETE (S3 + BD com rollback) ---------------------------------

// DeleteDocument godoc
// @Summary Remove um documento (S3 + BD) com rollback seguro
// @Tags Documentos
// @Produce json
// @Param id path string true "ID do documento"
// @Success 204
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /documents/{id} [delete]
func (h *DocumentHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do documento deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	lockKey := "doc:id:" + id.String()
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	doc, err := h.FindUC.Execute(ctx, id)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusNotFound, utils.Payload{Error: "not_found", Message: "Documento não encontrado"})
		return
	}

	// Verifica se o objeto existe — se não, apagamos apenas no BD
	exists, exErr := h.Storage.Exists(ctx, doc.ObjectKey)
	if exErr != nil {
		// Se não conseguimos checar, ainda tentamos backup (que tratará NoSuchKey)
	}

	// Criar backup (pode retornar no-ops se objeto não existir)
	_, restore, cleanup, bkErr := h.backupObject(ctx, doc.ObjectKey)
	if bkErr != nil {
		// Se o erro for “NoSuchKey”, seguimos sem backup; se for outro, falha
		if !isNoSuchKey(bkErr) {
			c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao preparar backup"})
			return
		}
		// caso NoSuchKey: restaura/cleanup virão no-ops
	}
	defer func() { _ = cleanup(ctx) }()

	// Se sabemos que existe, tentamos apagar; se não existe, ignoramos essa etapa
	if exists {
		if err := h.Storage.Delete(ctx, doc.ObjectKey); err != nil && !isNoSuchKey(err) {
			c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao apagar ficheiro no storage"})
			return
		}
	}

	// Remover no BD
	if err := h.DeleteUC.Execute(ctx, id); err != nil {
		// Tenta restaurar o ficheiro somente se ele existia e foi apagado nesta chamada
		if exists {
			_ = restore(ctx)
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao remover no BD; estado revertido"})
		return
	}

	// Sucesso: limpa backup (no-op se não houve backup)
	_ = cleanup(ctx)
	c.Status(http.StatusNoContent)
}

// --------------------------------- UPLOAD (S3 + BD com compensação reforçada) ---------------------------------

// UploadDocument godoc
// @Summary Upload de documento (S3 + BD)
// @Description Sobe ficheiro ao S3 e cria registo no BD.
// @Tags Documentos
// @Accept mpfd
// @Produce json
// @Param owner_type formData string true "ex.: employee, dependent"
// @Param owner_id formData string true "UUID do dono"
// @Param type formData string true "Tipo lógico do documento (BI, Contrato, etc.)"
// @Param file formData file true "Ficheiro"
// @Success 201 {object} dtos.DocumentResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 413 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /documents/upload [post]
func (h *DocumentHandler) Upload(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	// Limitar tamanho do corpo
	if h.MaxUploadSize > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.MaxUploadSize+1024)
	}

	var dto dtos.UploadDocumentForm
	if err := c.ShouldBindWith(&dto, binding.FormMultipart); err != nil {
		if fields := utils.HumanizeValidation(dto, err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação dos campos",
				Message: "Dados inválidos. Corrija os campos destacados.",
				Fields:  fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Ficheiro não fornecido"})
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Tamanho do ficheiro é zero"})
		return
	}
	if h.MaxUploadSize > 0 && header.Size > h.MaxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, utils.Payload{
			Error:   "payload_too_large",
			Message: "Insira um ficheiro com tamanho máximo de 2Mb",
			Fields:  []utils.FieldError{},
		})
		c.Abort()
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if len(h.AllowedExts) > 0 && !h.AllowedExts[ext] {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: fmt.Sprintf("Extensão '%s' não suportada", ext)})
		return
	}

	// Sniff content-type
	buf512 := make([]byte, 512)
	n, _ := file.Read(buf512)
	contentType := http.DetectContentType(buf512[:n])

	// Reset reader
	var rdr io.ReadSeeker
	if seeker, ok := file.(io.ReadSeeker); ok {
		rdr = seeker
		rdr.Seek(0, io.SeekStart)
	} else {
		mem := bytes.NewBuffer(nil)
		mem.Write(buf512[:n])
		if _, err := io.Copy(mem, file); err != nil {
			c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao ler ficheiro"})
			return
		}
		rdr = bytes.NewReader(mem.Bytes())
		header.Size = int64(mem.Len())
	}

	if len(h.AllowedTypes) > 0 && !h.AllowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: fmt.Sprintf("Content-Type '%s' não suportado", contentType)})
		return
	}

	ownerID, err := uuid.Parse(dto.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "owner_id", Label: "owner_id", Tag: "uuid", Message: "owner_id deve ser um UUID válido.", Value: dto.OwnerID},
			},
		})
		return
	}

	// __ADDED__: calcular MD5 e size (opcional para persistir/auditar)
	md5sum, size, rdr2, md5Err := computeMD5AndSize(rdr)
	if md5Err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao calcular MD5/tamanho"})
		return
	}
	rdr = rdr2

	// Lock por (ownerType + ownerID + filename) para evitar duplicidade simultânea
	lockKey := "doc:upload:" + strings.ToLower(dto.OwnerType) + ":" + ownerID.String() + ":" + utils.Normalize(header.Filename)
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	key := buildObjectKey(dto.OwnerType, ownerID, header.Filename)

	publicURL, err := h.Storage.Save(ctx, key, rdr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao subir ficheiro no storage"})
		return
	}

	input := documents_uc.CreateDocumentInput{
		OwnerType: dto.OwnerType,
		OwnerID:   ownerID,
		Type:      dto.Type,
		FileName:  sanitizeFilename(header.Filename),
		FileURL:   publicURL,
		Extension: strings.TrimPrefix(ext, "."),
		IsActive:  true,
		ObjectKey: key, // guarda a key real
		// Se existirem estes campos na tua entidade/DTO, podes preencher:
		// SizeBytes: size, // __ADDED__
		// MD5:       md5sum, // __ADDED__
		// MimeType:  contentType, // __ADDED__
	}
	doc, err := h.CreateUC.Execute(ctx, input)
	if err != nil {
		_ = h.Storage.Delete(ctx, key) // rollback do upload no S3
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao registar documento no BD; upload revertido"})
		return
	}

	// (Opcional) poderias devolver md5/size/mimetype também no DTO de resposta
	_ = size
	_ = md5sum
	_ = contentType

	c.JSON(http.StatusCreated, dtos.ToDocumentResponseDTO(doc))
}

// --------------------------------- DOWNLOAD (streaming) ---------------------------------

// DownloadDocument godoc
// @Summary Faz stream do documento armazenado (útil para objetos privados)
// @Tags Documentos
// @Produce octet-stream
// @Param id path string true "ID do documento"
// @Success 200
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /documents/{id}/download [get]
func (h *DocumentHandler) Download(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{ /* ... */ })
		return
	}

	doc, err := h.FindUC.Execute(ctx, id)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusNotFound, utils.Payload{Error: "not_found", Message: "Documento não encontrado"})
		return
	}

	rc, err := h.Storage.Open(ctx, doc.ObjectKey)
	if err != nil {
		if isNoSuchKey(err) {
			c.JSON(http.StatusNotFound, utils.Payload{Error: "not_found", Message: "Ficheiro não encontrado no armazenamento"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao abrir o ficheiro no storage"})
		return
	}
	defer rc.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, sanitizeFilename(doc.FileName.String())))
	c.Header("Content-Type", "application/octet-stream")
	_, _ = io.Copy(c.Writer, rc)
}

// --------------------------------- REPLACE FILE (S3 + BD com janela segura) ---------------------------------

// ReplaceDocumentFile godoc
// @Summary Substitui o ficheiro do documento (upload novo, atualiza BD e apaga o antigo com rollback)
// @Tags Documentos
// @Accept mpfd
// @Produce json
// @Param id path string true "ID do documento"
// @Param file formData file true "Novo ficheiro"
// @Success 200 {object} dtos.DocumentResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /documents/{id}/file [put]
// --------------------------------- REPLACE FILE (S3 + BD com janela segura e ObjectKey no BD) ---------------------------------

// ReplaceDocumentFile godoc
// @Summary Substitui o ficheiro do documento (upload novo, atualiza BD e apaga o antigo com rollback)
// @Tags Documentos
// @Accept mpfd
// @Produce json
// @Param id path string true "ID do documento"
// @Param file formData file true "Novo ficheiro"
// @Success 200 {object} dtos.DocumentResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /documents/{id}/file [put]
func (h *DocumentHandler) ReplaceFile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do documento deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	// Serializa a operação por documento
	lockKey := "doc:id:" + id.String()
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	// 1) Carrega estado atual do documento (inclui ObjectKey persistido)
	doc, err := h.FindUC.Execute(ctx, id)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusNotFound, utils.Payload{Error: "not_found", Message: "Documento não encontrado"})
		return
	}
	oldKey := doc.ObjectKey // << usar key guardada na BD

	// 2) Ler novo ficheiro
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Ficheiro não fornecido"})
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Tamanho do ficheiro é zero"})
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if len(h.AllowedExts) > 0 && !h.AllowedExts[ext] {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: fmt.Sprintf("Extensão '%s' não suportada", ext)})
		return
	}

	// Sniff content-type (apenas validação)
	buf512 := make([]byte, 512)
	n, _ := file.Read(buf512)
	contentType := http.DetectContentType(buf512[:n])

	fmt.Println("FASE 1")
	// Reset reader
	var rdr io.ReadSeeker
	if seeker, ok := file.(io.ReadSeeker); ok {
		rdr = seeker
		rdr.Seek(0, io.SeekStart)
	} else {
		mem := bytes.NewBuffer(nil)
		mem.Write(buf512[:n])
		if _, err := io.Copy(mem, file); err != nil {
			c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao ler ficheiro"})
			return
		}
		rdr = bytes.NewReader(mem.Bytes())
		header.Size = int64(mem.Len())
	}
	if len(h.AllowedTypes) > 0 && !h.AllowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: fmt.Sprintf("Content-Type '%s' não suportado", contentType)})
		return
	}

	// (Opcional) MD5/size do novo
	md5sum, size, rdr2, md5Err := computeMD5AndSize(rdr)
	if md5Err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao calcular MD5/tamanho"})
		return
	}
	rdr = rdr2

	fmt.Println("FASE 2")

	// 3) Upload do novo ficheiro (gera nova key lógica)
	newKey := buildObjectKey(doc.OwnerType.String(), doc.OwnerID, header.Filename)
	newURL, err := h.Storage.Save(ctx, newKey, rdr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao subir novo ficheiro"})
		return
	}

	// 4) Backup do antigo (pode ser no-op se o objeto antigo não existir)
	_, restoreOld, cleanupOld, bkErr := h.backupObject(ctx, oldKey)
	if bkErr != nil && !isNoSuchKey(bkErr) {
		// Falha a criar backup do antigo => desfaz upload do novo e aborta
		_ = h.Storage.Delete(ctx, newKey)
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao criar backup do ficheiro antigo"})
		return
	}
	defer func() { _ = cleanupOld(ctx) }()

	// 5) Atualizar BD para apontar ao novo ficheiro (inclui nova ObjectKey)
	updateDTO := dtos.UpdateDocumentDTO{
		Name:      sanitizeFilename(header.Filename),
		FileURL:   newURL,
		Extension: strings.TrimPrefix(ext, "."),
		ObjectKey: newKey, // << importante: substituir a key persistida!
		// (se existirem no DTO/entidade)
		// SizeBytes: size,
		// MD5:       md5sum,
		// MimeType:  contentType,
	}
	updated, updErr := h.UpdateUC.Execute(ctx, documents_uc.UpdateDocumentInput{
		ID:          doc.ID,
		DocumentDTO: updateDTO,
	})

	fmt.Println("UPDATED - ", updated.FileName)

	fmt.Println("UP ERROR - ", updErr)

	fmt.Println("FASE 3")

	if updErr != nil {
		// rollback: apaga novo e restaura antigo
		_ = h.Storage.Delete(ctx, newKey)
		_ = restoreOld(ctx) // no-op se antigo não existia
		if ok, payload, status := utils.HumanizeDB(updErr); ok {

			fmt.Println("STATUS - ", status)
			fmt.Println("PAYLOAD - ", payload)

			c.JSON(status, payload)
			return
		}
		fmt.Println("ERRRROOO - ", updErr)
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Falha ao atualizar registo no BD; versão antiga restaurada"})
		return
	}

	fmt.Println("FASE 4")
	// 6) Remover definitivamente o antigo (ignora NoSuchKey)
	if err := h.Storage.Delete(ctx, oldKey); err != nil && !isNoSuchKey(err) {
		// best-effort: poderias logar o erro aqui
	}

	_ = cleanupOld(ctx) // limpa backup (no-op se não houve backup)

	// (opcional) não usados diretamente:
	_ = size
	_ = md5sum
	_ = contentType

	fmt.Println("FASE 5")
	c.JSON(http.StatusOK, dtos.ToDocumentResponseDTO(updated))
}

// --------------------------------- HELPERS ---------------------------------

func buildObjectKey(ownerType string, ownerID uuid.UUID, originalFilename string) string {
	now := time.Now().UTC().Format("20060102T150405Z")
	base := sanitizeFilename(strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename)))
	ext := strings.ToLower(filepath.Ext(originalFilename))
	return fmt.Sprintf("%s/%s/%s_%s%s", strings.ToLower(ownerType), ownerID.String(), base, now, ext)
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "..", ".")
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_."
	var b strings.Builder
	for _, r := range name {
		if strings.ContainsRune(allowed, r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	s := b.String()
	if s == "" {
		return "file"
	}
	if len(s) > 100 {
		return s[:100]
	}
	return s
}

// Infere a key do objeto a partir da FileURL salva no BD.
// Funciona para URLs http(s) e também para s3://bucket/key
func objectKeyFromURL(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}
	// Em http(s) e s3://, a key fica no Path sem a primeira "/"
	return strings.TrimPrefix(u.Path, "/"), nil
}
