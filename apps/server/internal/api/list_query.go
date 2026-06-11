package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"robot-center/apps/server/internal/api/dto"
)

const operatorListMaxLimit = 200

type listQuery struct {
	filter   string
	hasLimit bool
	limit    int
	offset   int
	order    string
	sort     string
}

type listQueryResult[T any] struct {
	items []T
	meta  dto.ListResponseMeta
}

func parseListQuery(r *http.Request, allowedSorts map[string]string) listQuery {
	query := r.URL.Query()
	limit, hasLimit := listLimitQueryValue(query.Get("limit"))
	return listQuery{
		filter:   strings.TrimSpace(query.Get("filter")),
		hasLimit: hasLimit,
		limit:    limit,
		offset:   nonNegativeIntQueryValue(r, "offset", 0),
		order:    listOrderQueryValue(query.Get("order")),
		sort:     listSortQueryValue(query.Get("sort"), allowedSorts),
	}
}

func applyListQuery[T any](
	items []T,
	query listQuery,
	matchesFilter func(T, string) bool,
	lessBySort func(left T, right T, sortKey string) bool,
) listQueryResult[T] {
	filteredItems := make([]T, 0, len(items))
	for _, item := range items {
		if query.filter == "" || matchesFilter == nil || matchesFilter(item, query.filter) {
			filteredItems = append(filteredItems, item)
		}
	}

	if query.sort != "" && lessBySort != nil {
		sort.SliceStable(filteredItems, func(leftIndex int, rightIndex int) bool {
			if query.order == "desc" {
				return lessBySort(filteredItems[rightIndex], filteredItems[leftIndex], query.sort)
			}
			return lessBySort(filteredItems[leftIndex], filteredItems[rightIndex], query.sort)
		})
	}

	total := len(filteredItems)
	offset := query.offset
	if offset > total {
		offset = total
	}

	limit := total
	if query.hasLimit {
		limit = query.limit
	}

	end := total
	if query.hasLimit && offset+limit < end {
		end = offset + limit
	}

	pageItems := filteredItems[offset:end]
	return listQueryResult[T]{
		items: pageItems,
		meta: dto.ListResponseMeta{
			Page: dto.ListPageResponse{
				Limit:    limit,
				Offset:   offset,
				Total:    total,
				Returned: len(pageItems),
				HasMore:  end < total,
			},
			Query: dto.ListQueryResponse{
				Filter: query.filter,
				Order:  query.order,
				Sort:   query.sort,
			},
		},
	}
}

func listLimitQueryValue(rawValue string) (int, bool) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return 0, false
	}
	value, err := strconv.Atoi(rawValue)
	if err != nil || value <= 0 {
		return operatorListMaxLimit, true
	}
	return clampIntValue(value, 1, operatorListMaxLimit), true
}

func listOrderQueryValue(rawValue string) string {
	switch strings.ToLower(strings.TrimSpace(rawValue)) {
	case "desc":
		return "desc"
	default:
		return "asc"
	}
}

func listSortQueryValue(rawValue string, allowedSorts map[string]string) string {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return ""
	}
	if canonicalSort, ok := allowedSorts[rawValue]; ok {
		return canonicalSort
	}
	if canonicalSort, ok := allowedSorts[strings.ToLower(rawValue)]; ok {
		return canonicalSort
	}
	return ""
}

func containsListFilterValue(filter string, values ...string) bool {
	normalizedFilter := strings.ToLower(strings.TrimSpace(filter))
	if normalizedFilter == "" {
		return true
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), normalizedFilter) {
			return true
		}
	}
	return false
}

func lessListString(left string, right string) bool {
	return strings.ToLower(left) < strings.ToLower(right)
}

func lessListTime(left time.Time, right time.Time) bool {
	if left.Equal(right) {
		return false
	}
	return left.Before(right)
}

func lessOptionalListTime(left *time.Time, right *time.Time) bool {
	if left == nil && right == nil {
		return false
	}
	if left == nil {
		return true
	}
	if right == nil {
		return false
	}
	return lessListTime(*left, *right)
}
