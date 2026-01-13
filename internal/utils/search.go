package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchInput struct {
	SearchText string
	Filter     string
	Limit      int
	Offset     int
}

// applyDefaults padroniza os valores de paginação
func ApplyDefaults(input *SearchInput) {
	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.Offset < 0 {
		input.Offset = 0
	}
}

func ParseSearchInput(c *gin.Context) SearchInput {
	search := strings.TrimSpace(c.Query("search"))
	filter := strings.TrimSpace(c.Query("filter"))
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	return SearchInput{
		SearchText: search,
		Filter:     filter,
		Limit:      limit,
		Offset:     offset,
	}
}
