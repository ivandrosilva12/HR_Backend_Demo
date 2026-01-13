-- Extensões necessárias
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =========================
-- 1. Tipos (ENUMs)
-- =========================
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'position_tipo') THEN
    CREATE TYPE position_tipo AS ENUM ('employee','boss');
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'worker_status') THEN
    CREATE TYPE worker_status AS ENUM ('activo','inactivo');
  END IF;
END $$;

-- =========================
-- 2. Tabelas base
-- =========================

-- 2.1 Áreas de Estudo
CREATE TABLE areas_estudo (
    id UUID PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.2 Províncias
CREATE TABLE provinces (
    id UUID PRIMARY KEY,
    nome VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.3 Departamentos (com hierarquia)
CREATE TABLE departments (
    id UUID PRIMARY KEY,
    nome VARCHAR(100) NOT NULL UNIQUE,
    parent_id UUID NULL REFERENCES departments(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_departments_parent ON departments(parent_id);

-- 2.4 Posições (Cargos)
CREATE TABLE positions (
    id UUID PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    department_id UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    tipo position_tipo NOT NULL DEFAULT 'employee',
    max_headcount INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (nome, department_id),
    CONSTRAINT chk_positions_max_headcount CHECK (max_headcount >= 1)
);
CREATE INDEX IF NOT EXISTS idx_positions_department ON positions(department_id);

-- 2.5 Municípios
CREATE TABLE municipalities (
    id UUID PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    province_id UUID NOT NULL REFERENCES provinces(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.6 Distritos
CREATE TABLE districts (
    id UUID PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    municipio_id UUID NOT NULL REFERENCES municipalities(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.7 Funcionários
CREATE TABLE employees (
    id UUID PRIMARY KEY,
    employee_number INTEGER NOT NULL UNIQUE,
    full_name VARCHAR(100) NOT NULL,
    gender VARCHAR(10) NOT NULL,
    date_of_birth DATE NOT NULL,
    nationality VARCHAR(50) NOT NULL,
    marital_status VARCHAR(20) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(100) NOT NULL,
    bi VARCHAR(14) NOT NULL,
    id_date  DATE NOT NULL,
    iban VARCHAR(34) NOT NULL,
    department_id UUID NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE RESTRICT,
    address VARCHAR(200) NOT NULL,
    district_id UUID NOT NULL REFERENCES districts(id) ON DELETE RESTRICT,
    hiring_date DATE NOT NULL,
    contract_type VARCHAR(20) NOT NULL,
    salary NUMERIC(12, 2) NOT NULL,
    social_security VARCHAR(12) NOT NULL,
    -- supervisor_id REMOVIDO
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.8 Documentos
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    owner_type VARCHAR(20) NOT NULL,
    owner_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    file_name TEXT NOT NULL,
    file_url TEXT NOT NULL,
    extension VARCHAR(10) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    object_key TEXT NOT NULL
);

-- 2.9 Estado dos Funcionários
CREATE TABLE employee_statuses (
    id UUID PRIMARY KEY,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL,
    reason VARCHAR(100) NOT NULL,
    observacoes TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.10 Dependentes
CREATE TABLE dependents (
    id UUID PRIMARY KEY,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    full_name VARCHAR(100) NOT NULL,
    relationship VARCHAR(30) NOT NULL,
    gender VARCHAR(10) NOT NULL,
    date_of_birth DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.11 Histórico Educacional
CREATE TABLE education_histories (
    id UUID PRIMARY KEY,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    institution TEXT NOT NULL,
    degree TEXT NOT NULL,
    field_of_study TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.12 Histórico Profissional
CREATE TABLE work_histories (
    id UUID PRIMARY KEY,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    company TEXT NOT NULL,
    "position" TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    responsibilities TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.13 Histórico de Vínculos (substitui supervisor_histories)
CREATE TABLE worker_histories (
    id UUID PRIMARY KEY,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE,
    status worker_status NOT NULL DEFAULT 'activo',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =========================
-- 3. Identidade & Ajustes
-- =========================

-- Tornar employee_number auto-gerado (IDENTITY)
ALTER TABLE employees
  ALTER COLUMN employee_number
  ADD GENERATED ALWAYS AS IDENTITY;

-- Ajustar o próximo valor para não colidir com existentes (IDENTITY-friendly)
DO $$
DECLARE
  nextval BIGINT;
BEGIN
  SELECT COALESCE(MAX(employee_number), 0) + 1 INTO nextval FROM public.employees;
  EXECUTE format('ALTER TABLE public.employees ALTER COLUMN employee_number RESTART WITH %s', nextval);
END $$;

-- =========================
-- 4. Índices
-- =========================
CREATE INDEX idx_work_employee ON work_histories (employee_id);

CREATE INDEX idx_worker_employee ON worker_histories (employee_id);
CREATE INDEX idx_worker_position ON worker_histories (position_id);
CREATE INDEX idx_worker_active ON worker_histories (employee_id, status, end_date);

CREATE INDEX idx_employee_status_employee ON employee_statuses (employee_id);
CREATE INDEX idx_employee_status_current ON employee_statuses (employee_id, is_current);
CREATE INDEX idx_education_employee ON education_histories (employee_id);
CREATE INDEX idx_dependents_employee ON dependents (employee_id);

CREATE INDEX idx_documents_owner ON documents (owner_type, owner_id);
CREATE INDEX idx_documents_type ON documents (type);
CREATE INDEX idx_documents_ownerid_uploadedat ON documents (owner_id, uploaded_at DESC);
CREATE INDEX idx_documents_owner_type_date ON documents (owner_id, type, uploaded_at DESC);
CREATE UNIQUE INDEX idx_documents_object_key ON documents(object_key);

CREATE UNIQUE INDEX idx_employees_bi ON employees (bi);
CREATE UNIQUE INDEX idx_employees_email ON employees (LOWER(email));
CREATE INDEX idx_employees_department_id ON employees (department_id);
CREATE INDEX idx_employees_position_id ON employees (position_id);
CREATE INDEX idx_employees_district_id ON employees (district_id);

CREATE UNIQUE INDEX unique_district_per_municipality ON districts (LOWER(nome), municipio_id);
CREATE UNIQUE INDEX unique_municipality_per_province ON municipalities (LOWER(nome), province_id);

-- =========================
-- 5. Triggers utilitários
-- =========================

-- 5.1 updated_at automático
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'areas_estudo', 'provinces', 'departments', 'positions',
        'municipalities', 'districts', 'employees',
        'employee_statuses', 'dependents', 'education_histories',
        'work_histories', 'worker_histories'
    ])
    LOOP
        EXECUTE format('
            CREATE TRIGGER trg_set_updated_at_%I
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column()', tbl, tbl);
    END LOOP;
END $$;

-- 5.2 Evitar ciclos na hierarquia de departamentos
CREATE OR REPLACE FUNCTION prevent_department_cycle()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.parent_id IS NULL THEN
    RETURN NEW;
  END IF;

  IF NEW.parent_id = NEW.id THEN
    RAISE EXCEPTION 'Departamento não pode ser pai de si próprio' USING ERRCODE='23514';
  END IF;

  IF EXISTS (
    WITH RECURSIVE chain(id) AS (
      SELECT NEW.parent_id
      UNION ALL
      SELECT d.parent_id FROM departments d JOIN chain ON d.id = chain.id
      WHERE d.parent_id IS NOT NULL
    )
    SELECT 1 FROM chain WHERE id = NEW.id
  ) THEN
    RAISE EXCEPTION 'Ciclo detectado na hierarquia de departamentos' USING ERRCODE='23514';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_department_cycle ON departments;
CREATE TRIGGER trg_department_cycle
BEFORE INSERT OR UPDATE OF parent_id ON departments
FOR EACH ROW EXECUTE FUNCTION prevent_department_cycle();

-- 5.3 Bloqueia a remoção de um departamento que tenha filhos (mensagem amigável)
CREATE OR REPLACE FUNCTION prevent_delete_parent_department()
RETURNS TRIGGER AS $$
BEGIN
  IF EXISTS (SELECT 1 FROM departments d WHERE d.parent_id = OLD.id) THEN
    RAISE EXCEPTION 'Não é permitido apagar o departamento %, pois ele é pai de outros.',
      OLD.id USING ERRCODE='23503';
  END IF;
  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_delete_parent_department ON departments;
CREATE TRIGGER trg_prevent_delete_parent_department
BEFORE DELETE ON departments
FOR EACH ROW
EXECUTE FUNCTION prevent_delete_parent_department();

-- 5.4 Regras em Employees: coerência de depto/posição e capacidade
CREATE OR REPLACE FUNCTION enforce_position_rules()
RETURNS TRIGGER AS $$
DECLARE
  cap INT;
  pos_dept UUID;
  current_count INT;
BEGIN
  IF NEW.position_id IS NULL THEN
    RETURN NEW;
  END IF;

  SELECT max_headcount, department_id
    INTO cap, pos_dept
  FROM positions
  WHERE id = NEW.position_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Posição inexistente: %', NEW.position_id USING ERRCODE='23503';
  END IF;

  IF NEW.department_id IS DISTINCT FROM pos_dept THEN
    RAISE EXCEPTION 'Departamento do funcionário (%) difere do departamento da posição (%)',
      NEW.department_id, pos_dept USING ERRCODE='23514';
  END IF;

  SELECT COUNT(*) INTO current_count
  FROM employees e
  WHERE e.position_id = NEW.position_id
    AND e.is_active = TRUE
    AND (TG_OP = 'INSERT' OR e.id <> NEW.id);

  IF NEW.is_active AND current_count >= cap THEN
    RAISE EXCEPTION 'Capacidade esgotada para a posição (max=%)', cap USING ERRCODE='23514';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_enforce_position_rules ON employees;
CREATE TRIGGER trg_enforce_position_rules
BEFORE INSERT OR UPDATE OF position_id, department_id, is_active
ON employees
FOR EACH ROW EXECUTE FUNCTION enforce_position_rules();

-- 5.5 Histórico de vínculos automático (worker_histories)
-- Ao inserir funcionário → cria histórico activo (sem end_date)
CREATE OR REPLACE FUNCTION worker_history_on_insert()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO worker_histories (id, employee_id, position_id, start_date, status, created_at, updated_at)
  VALUES (
    gen_random_uuid(),
    NEW.id,
    NEW.position_id,
    COALESCE(NEW.hiring_date, CURRENT_DATE),
    'activo',
    NOW(), NOW()
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_worker_history_after_insert') THEN
    CREATE TRIGGER trg_worker_history_after_insert
    AFTER INSERT ON employees
    FOR EACH ROW
    EXECUTE FUNCTION worker_history_on_insert();
  END IF;
END $$;

-- Ao alterar posição → fecha histórico anterior e abre novo
CREATE OR REPLACE FUNCTION worker_history_on_position_change()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.position_id IS DISTINCT FROM OLD.position_id THEN
    UPDATE worker_histories
       SET end_date = COALESCE(NEW.hiring_date, CURRENT_DATE),
           status   = 'inactivo',
           updated_at = NOW()
     WHERE employee_id = NEW.id
       AND end_date IS NULL;

    INSERT INTO worker_histories (id, employee_id, position_id, start_date, status, created_at, updated_at)
    VALUES (
      gen_random_uuid(),
      NEW.id,
      NEW.position_id,
      COALESCE(NEW.hiring_date, CURRENT_DATE),
      'activo',
      NOW(), NOW()
    );
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_worker_history_after_update') THEN
    CREATE TRIGGER trg_worker_history_after_update
    AFTER UPDATE OF position_id ON employees
    FOR EACH ROW
    EXECUTE FUNCTION worker_history_on_position_change();
  END IF;
END $$;

-- =========================
-- 6. Views
-- =========================

-- 6.1 Ocupação atual por posição (current_headcount/remaining)
CREATE OR REPLACE VIEW positions_with_occupancy AS
SELECT
  p.id,
  p.nome,
  p.department_id,
  p.max_headcount,
  COUNT(e.id) FILTER (WHERE e.is_active) AS current_headcount,
  (p.max_headcount - COUNT(e.id) FILTER (WHERE e.is_active)) AS remaining
FROM positions p
LEFT JOIN employees e ON e.position_id = p.id
GROUP BY p.id;

-- 6.2 Municípios (para pesquisa)
CREATE OR REPLACE VIEW view_municipalities_search AS
SELECT 
    m.id,
    m.nome,
    m.province_id,
    p.nome AS province_nome,
    m.created_at,
    m.updated_at
FROM municipalities m
JOIN provinces p ON m.province_id = p.id;

-- 6.3 Posições (para pesquisa)
CREATE OR REPLACE VIEW view_positions_search AS
SELECT 
    p.id,
    p.nome,
    p.department_id,
    d.nome AS department_name,
    p.created_at,
    p.updated_at
FROM positions p
JOIN departments d ON p.department_id = d.id;

-- 6.4 Árvore de departamentos (com nº de ativos)
CREATE OR REPLACE VIEW departments_tree AS
WITH RECURSIVE t AS (
  SELECT d.id, d.nome, d.parent_id, 0 AS depth
  FROM departments d
  WHERE d.parent_id IS NULL
  UNION ALL
  SELECT c.id, c.nome, c.parent_id, t.depth + 1
  FROM departments c
  JOIN t ON c.parent_id = t.id
)
SELECT
  t.*,
  (SELECT COUNT(*) FROM employees e WHERE e.department_id = t.id AND e.is_active) AS active_employees
FROM t
ORDER BY depth, nome;

-- =========================
-- 7. Funções de Pesquisa
-- =========================

-- 7.1 Municípios
CREATE OR REPLACE FUNCTION search_municipalities(
    search_text TEXT DEFAULT '',
    province_filter TEXT DEFAULT '',
    limit_results INT DEFAULT 10,
    offset_results INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    name TEXT,
    province_id UUID,
    province_nome TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.id,
        m.nome::text,
        m.province_id,
        p.nome::text,
        m.created_at,
        m.updated_at
    FROM municipalities m
    JOIN provinces p ON m.province_id = p.id
    WHERE 
        (search_text = '' OR LOWER(m.nome) LIKE '%' || LOWER(search_text) || '%') AND
        (province_filter = '' OR LOWER(p.nome) = LOWER(province_filter))
    ORDER BY m.created_at DESC
    LIMIT limit_results OFFSET offset_results;
END;
$$ LANGUAGE plpgsql;

-- 7.2 Departamentos
CREATE OR REPLACE FUNCTION search_departments(
    search_text TEXT DEFAULT '',
    limit_results INT DEFAULT 10,
    offset_results INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    name TEXT,
    parent_id UUID,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.id,
        d.nome::text,
        d.parent_id,
        d.created_at,
        d.updated_at
    FROM departments d
    WHERE (search_text = '' OR LOWER(d.nome) LIKE '%' || LOWER(search_text) || '%')
    ORDER BY d.created_at DESC
    LIMIT limit_results OFFSET offset_results;
END;
$$ LANGUAGE plpgsql;

-- 7.3 Funcionários
CREATE OR REPLACE FUNCTION search_employees(
    search_text TEXT,
    department_filter TEXT,
    limit_val INTEGER,
    offset_val INTEGER
)
RETURNS TABLE (
    id UUID,
    employee_number INTEGER,
    full_name TEXT,
    gender TEXT,
    date_of_birth DATE,
    nationality TEXT,
    marital_status TEXT,
    phone_number TEXT,
    email TEXT,
    bi TEXT,
    id_date DATE,
    iban TEXT,
    department_id UUID,
    position_id UUID,
    address TEXT,
    district_id UUID,
    hiring_date DATE,
    contract_type TEXT,
    salary NUMERIC,
    social_security TEXT,
    is_active BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.id,
        e.employee_number,
        e.full_name::text,
        e.gender::text,
        e.date_of_birth,
        e.nationality::text,
        e.marital_status::text,
        e.phone_number::text,
        e.email::text,
        e.bi::text,
        e.id_date,
        e.iban::text,
        e.department_id,
        e.position_id,
        e.address::text,
        e.district_id,
        e.hiring_date,
        e.contract_type::text,
        e.salary,
        e.social_security::text,
        e.is_active,
        e.created_at,
        e.updated_at
    FROM employees e
    WHERE (
        LOWER(e.full_name) LIKE '%' || LOWER(search_text) || '%' OR
        LOWER(e.email) LIKE '%' || LOWER(search_text) || '%' OR
        LOWER(e.phone_number) LIKE '%' || LOWER(search_text) || '%' OR
        LOWER(e.bi) LIKE '%' || LOWER(search_text) || '%'
    )
    AND (
        department_filter IS NULL OR department_filter = '' OR
        EXISTS (
            SELECT 1 FROM departments d
            WHERE d.id = e.department_id AND LOWER(d.nome) LIKE '%' || LOWER(department_filter) || '%'
        )
    )
    ORDER BY e.created_at DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.4 Documentos
CREATE OR REPLACE FUNCTION search_documents(
    search_text TEXT DEFAULT '',
    type_filter TEXT DEFAULT '',
    owner_type_filter TEXT DEFAULT '',
    limit_val INT DEFAULT 10,
    offset_val INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    owner_type TEXT,
    owner_id UUID,
    type TEXT,
    file_name TEXT,
    file_url TEXT,
    extension TEXT,
    is_active BOOLEAN,
    uploaded_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        id,
        owner_type,
        owner_id,
        type,
        file_name,
        file_url,
        extension,
        is_active,
        uploaded_at
    FROM documents
    WHERE
        (search_text = '' OR LOWER(file_name) LIKE '%' || LOWER(search_text) || '%')
        AND (type_filter = '' OR LOWER(type) = LOWER(type_filter))
        AND (owner_type_filter = '' OR LOWER(owner_type) = LOWER(owner_type_filter))
    ORDER BY uploaded_at DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.5 Estados dos funcionários
CREATE OR REPLACE FUNCTION search_employee_statuses(
    employee_filter UUID DEFAULT NULL,
    status_filter TEXT DEFAULT '',
    limit_val INT DEFAULT 10,
    offset_val INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    employee_id UUID,
    status TEXT,
    reason TEXT,
    observacoes TEXT,
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.id, s.employee_id, s.status, s.reason, s.observacoes, s.start_date, s.end_date,
        s.is_current, s.created_at, s.updated_at
    FROM employee_statuses s
    WHERE
        (employee_filter IS NULL OR s.employee_id = employee_filter)
        AND (status_filter = '' OR LOWER(s.status) = LOWER(status_filter))
    ORDER BY s.start_date DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.6 Dependentes
CREATE OR REPLACE FUNCTION search_dependents(
    search_text TEXT DEFAULT '',
    employee_filter UUID DEFAULT NULL,
    limit_val INT DEFAULT 10,
    offset_val INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    employee_id UUID,
    full_name TEXT,
    relationship TEXT,
    gender TEXT,
    date_of_birth DATE,
    is_active BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.id, d.employee_id, d.full_name, d.relationship, d.gender,
        d.date_of_birth, d.is_active, d.created_at, d.updated_at
    FROM dependents d
    WHERE 
        (search_text = '' OR LOWER(d.full_name) LIKE '%' || LOWER(search_text) || '%')
        AND (employee_filter IS NULL OR d.employee_id = employee_filter)
    ORDER BY d.created_at DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.7 Histórico Educacional
CREATE OR REPLACE FUNCTION search_education_histories(
    employee_filter UUID DEFAULT NULL,
    search_text TEXT DEFAULT '',
    start_date_filter DATE DEFAULT NULL,
    end_date_filter DATE DEFAULT NULL,
    limit_val INT DEFAULT 10,
    offset_val INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    employee_id UUID,
    institution TEXT,
    degree TEXT,
    field_of_study TEXT,
    start_date DATE,
    end_date DATE,
    description TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        e.id, e.employee_id, e.institution, e.degree, e.field_of_study,
        e.start_date, e.end_date, e.description, e.created_at, e.updated_at
    FROM education_histories e
    WHERE
        (search_text = '' OR LOWER(e.institution) LIKE '%' || LOWER(search_text) || '%')
        AND (employee_filter IS NULL OR e.employee_id = employee_filter)
        AND (start_date_filter IS NULL OR e.start_date >= start_date_filter)
        AND (end_date_filter IS NULL OR e.end_date <= end_date_filter)
    ORDER BY e.start_date DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.8 Histórico Profissional
CREATE OR REPLACE FUNCTION search_work_histories(
    employee_filter UUID DEFAULT NULL,
    search_text TEXT DEFAULT '',
    start_date_filter DATE DEFAULT NULL,
    end_date_filter DATE DEFAULT NULL,
    limit_val INT DEFAULT 10,
    offset_val INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    employee_id UUID,
    company TEXT,
    "position" TEXT,
    start_date DATE,
    end_date DATE,
    responsibilities TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        w.id, w.employee_id, w.company, w."position", w.start_date, w.end_date,
        w.responsibilities, w.created_at, w.updated_at
    FROM work_histories w
    WHERE
        (search_text = '' OR LOWER(w.company) LIKE '%' || LOWER(search_text) || '%')
        AND (employee_filter IS NULL OR w.employee_id = employee_filter)
        AND (start_date_filter IS NULL OR w.start_date >= start_date_filter)
        AND (end_date_filter IS NULL OR w.end_date <= end_date_filter)
    ORDER BY w.start_date DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.9 Histórico de vínculos (worker_histories)
CREATE OR REPLACE FUNCTION search_worker_histories(
    employee_filter UUID DEFAULT NULL,
    position_filter UUID DEFAULT NULL,
    status_filter TEXT DEFAULT '',
    start_date_filter DATE DEFAULT NULL,
    end_date_filter DATE DEFAULT NULL,
    limit_val INT DEFAULT 10,
    offset_val INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    employee_id UUID,
    position_id UUID,
    status TEXT,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        wh.id, wh.employee_id, wh.position_id, wh.status::text,
        wh.start_date, wh.end_date, wh.created_at, wh.updated_at
    FROM worker_histories wh
    WHERE
        (employee_filter IS NULL OR wh.employee_id = employee_filter) AND
        (position_filter IS NULL OR wh.position_id = position_filter) AND
        (status_filter = '' OR LOWER(wh.status::text) = LOWER(status_filter)) AND
        (start_date_filter IS NULL OR wh.start_date >= start_date_filter) AND
        (end_date_filter IS NULL OR wh.end_date <= end_date_filter)
    ORDER BY wh.start_date DESC
    LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql;

-- 7.10 Posições
CREATE OR REPLACE FUNCTION search_positions(
    search_text TEXT DEFAULT '',
    department_filter TEXT DEFAULT '',
    limit_results INT DEFAULT 10,
    offset_results INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    nome TEXT,
    department_id UUID,
    department_name TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p.id,
        p.nome::text,
        p.department_id,
        d.nome::text,
        p.created_at,
        p.updated_at
    FROM positions p
    JOIN departments d ON p.department_id = d.id
    WHERE 
        (search_text = '' OR LOWER(p.nome) LIKE '%' || LOWER(search_text) || '%')
        AND (department_filter = '' OR LOWER(d.nome) = LOWER(department_filter))
    ORDER BY p.created_at DESC
    LIMIT limit_results OFFSET offset_results;
END;
$$ LANGUAGE plpgsql;

-- 7.11 Distritos
CREATE OR REPLACE FUNCTION search_districts(
    search_text TEXT DEFAULT '',
    municipio_filter TEXT DEFAULT '',
    limit_results INT DEFAULT 10,
    offset_results INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    nome TEXT,
    municipio_id UUID,
    municipio_name TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.id,
        d.nome::text,
        d.municipio_id,
        m.nome::text,
        d.created_at,
        d.updated_at
    FROM districts d
    JOIN municipalities m ON d.municipio_id = m.id
    WHERE
        (search_text = '' OR LOWER(d.nome) LIKE '%' || LOWER(search_text) || '%')
        AND (municipio_filter = '' OR LOWER(m.nome) = LOWER(municipio_filter))
    ORDER BY d.created_at DESC
    LIMIT limit_results OFFSET offset_results;
END;
$$ LANGUAGE plpgsql;

-- 7.12 Províncias
CREATE OR REPLACE FUNCTION search_provinces(
    search_text TEXT DEFAULT '',
    limit_results INT DEFAULT 10,
    offset_results INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    name TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p.id,
        p.nome::text,
        p.created_at,
        p.updated_at
    FROM provinces p
    WHERE (search_text = '' OR LOWER(p.nome) LIKE '%' || LOWER(search_text) || '%')
    ORDER BY p.created_at DESC
    LIMIT limit_results OFFSET offset_results;
END;
$$ LANGUAGE plpgsql;

-- 7.13 Áreas de estudo
CREATE OR REPLACE FUNCTION search_areas_estudo(
    search_text TEXT DEFAULT '',
    limit_results INT DEFAULT 10,
    offset_results INT DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    name TEXT,
    description TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.nome::text,
        a.description::text,
        a.created_at,
        a.updated_at
    FROM areas_estudo a
    WHERE (search_text = '' OR LOWER(a.nome) LIKE '%' || LOWER(search_text) || '%')
    ORDER BY a.created_at DESC
    LIMIT limit_results OFFSET offset_results;
END;
$$ LANGUAGE plpgsql;

-- 7.14 Lista de funcionários por departamento (com subárvore opcional)
CREATE OR REPLACE FUNCTION list_employees_by_department(
  department_root UUID,
  include_children BOOLEAN DEFAULT TRUE,
  only_active BOOLEAN DEFAULT TRUE,
  limit_val INT DEFAULT 50,
  offset_val INT DEFAULT 0
)
RETURNS TABLE (
  id UUID,
  employee_number INTEGER,
  full_name TEXT,
  email TEXT,
  phone_number TEXT,
  department_id UUID,
  position_id UUID,
  hiring_date DATE,
  is_active BOOLEAN,
  created_at TIMESTAMP
) AS $$
BEGIN
  RETURN QUERY
  WITH RECURSIVE depts AS (
    SELECT department_root AS id
    UNION ALL
    SELECT d.id FROM departments d JOIN depts ON d.parent_id = depts.id
  )
  SELECT
    e.id, e.employee_number, e.full_name, e.email, e.phone_number,
    e.department_id, e.position_id, e.hiring_date, e.is_active, e.created_at
  FROM employees e
  WHERE
    e.department_id IN (
      SELECT CASE WHEN include_children THEN id ELSE department_root END FROM depts
    )
    AND (only_active IS FALSE OR e.is_active)
  ORDER BY e.created_at DESC
  LIMIT limit_val OFFSET offset_val;
END;
$$ LANGUAGE plpgsql STABLE;

-- Totais para 1 departamento (com opção de incluir a subárvore)
CREATE OR REPLACE FUNCTION department_position_totals(
  department_root   UUID,
  include_children  BOOLEAN DEFAULT FALSE
)
RETURNS TABLE (
  department_id      UUID,
  department_nome    TEXT,
  total_positions    INTEGER,
  occupied_positions INTEGER,
  available_positions INTEGER
) AS $$
  WITH RECURSIVE depts AS (
    SELECT department_root AS id
    UNION ALL
    SELECT d.id
    FROM departments d
    JOIN depts ON d.parent_id = depts.id
  ),
  scope AS (
    SELECT CASE WHEN include_children THEN id ELSE department_root END AS id
    FROM depts
  ),
  agg AS (
    SELECT
      p.department_id,
      SUM(p.max_headcount)::int     AS total_positions,
      SUM(p.current_headcount)::int AS occupied_positions
    FROM positions_with_occupancy p
    WHERE p.department_id IN (SELECT id FROM scope)
    GROUP BY p.department_id
  )
  SELECT
    d.id,
    d.nome,
    COALESCE(a.total_positions, 0)                              AS total_positions,
    COALESCE(a.occupied_positions, 0)                           AS occupied_positions,
    COALESCE(a.total_positions, 0) - COALESCE(a.occupied_positions, 0) AS available_positions
  FROM departments d
  LEFT JOIN agg a ON a.department_id = d.id
  WHERE d.id IN (SELECT id FROM scope)
  ORDER BY d.nome;
$$ LANGUAGE sql STABLE;