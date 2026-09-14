package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination_NewPagination(t *testing.T) {
	t.Parallel()

	t.Run("success - valid page and perPage", func(t *testing.T) {
		p := NewPagination(1, 20)

		assert.Equal(t, int64(1), p.page)
		assert.Equal(t, int64(20), p.perPage)
		assert.Equal(t, int64(0), p.totalCount)
	})

	t.Run("success - page 0 sets to 1", func(t *testing.T) {
		p := NewPagination(0, 20)

		assert.Equal(t, int64(1), p.page)
		assert.Equal(t, int64(20), p.perPage)
	})

	t.Run("success - negative page sets to 1", func(t *testing.T) {
		p := NewPagination(-5, 20)

		assert.Equal(t, int64(1), p.page)
		assert.Equal(t, int64(20), p.perPage)
	})

	t.Run("success - invalid perPage sets to 20", func(t *testing.T) {
		p := NewPagination(1, 100)

		assert.Equal(t, int64(1), p.page)
		assert.Equal(t, int64(20), p.perPage)
	})

	t.Run("success - perPage 50", func(t *testing.T) {
		p := NewPagination(1, 50)

		assert.Equal(t, int64(1), p.page)
		assert.Equal(t, int64(50), p.perPage)
	})
}

func TestPagination_GetPage(t *testing.T) {
	t.Parallel()

	t.Run("success - totalCount 0 returns 1", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(0)

		page := p.GetPage()

		assert.Equal(t, int64(1), page)
	})

	t.Run("success - page within range", func(t *testing.T) {
		p := NewPagination(2, 20)
		p.SetTotalCount(100)

		page := p.GetPage()

		assert.Equal(t, int64(2), page)
	})

	t.Run("success - page greater than totalPages returns last page", func(t *testing.T) {
		p := NewPagination(10, 20)
		p.SetTotalCount(100)

		page := p.GetPage()

		assert.Equal(t, int64(5), page)
	})

	t.Run("success - page 1 with totalCount 5", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(5)

		page := p.GetPage()

		assert.Equal(t, int64(1), page)
	})
}

func TestPagination_GetOffset(t *testing.T) {
	t.Parallel()

	t.Run("success - page 1 offset 0", func(t *testing.T) {
		p := NewPagination(1, 20)

		offset := p.GetOffset()

		assert.Equal(t, int64(0), offset)
	})

	t.Run("success - page 2 offset 20", func(t *testing.T) {
		p := NewPagination(2, 20)
		p.SetTotalCount(100)

		offset := p.GetOffset()

		assert.Equal(t, int64(20), offset)
	})

	t.Run("success - page 3 perPage 50 offset 100", func(t *testing.T) {
		p := NewPagination(3, 50)
		p.SetTotalCount(700)

		offset := p.GetOffset()

		assert.Equal(t, int64(100), offset)
	})

	t.Run("success - page 0 offset 0", func(t *testing.T) {
		p := NewPagination(0, 20)
		p.SetTotalCount(100)

		offset := p.GetOffset()

		assert.Equal(t, int64(0), offset)
	})
}

func TestPagination_GetTotalPages(t *testing.T) {
	t.Parallel()

	t.Run("success - totalCount 0 returns 0", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(0)

		totalPages := p.GetTotalPages()

		assert.Equal(t, int64(0), totalPages)
	})

	t.Run("success - totalCount 50 perPage 20 returns 3", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(50)

		totalPages := p.GetTotalPages()

		assert.Equal(t, int64(3), totalPages)
	})

	t.Run("success - totalCount 40 perPage 20 returns 2", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(40)

		totalPages := p.GetTotalPages()

		assert.Equal(t, int64(2), totalPages)
	})

	t.Run("success - totalCount 20 perPage 20 returns 1", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(20)

		totalPages := p.GetTotalPages()

		assert.Equal(t, int64(1), totalPages)
	})

	t.Run("success - totalCount 5 perPage 20 returns 1", func(t *testing.T) {
		p := NewPagination(1, 20)
		p.SetTotalCount(5)

		totalPages := p.GetTotalPages()

		assert.Equal(t, int64(1), totalPages)
	})

	t.Run("success - totalCount 100 perPage 50 returns 2", func(t *testing.T) {
		p := NewPagination(1, 50)
		p.SetTotalCount(100)

		totalPages := p.GetTotalPages()

		assert.Equal(t, int64(2), totalPages)
	})
}
