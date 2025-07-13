package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/google/uuid"
	"github.com/satrunjis/user-service/internal/domain"
	"github.com/satrunjis/user-service/internal/service"
)

const usersIndex = "users"
const mappings = `{
  "mappings": {
    "properties": {
      "login":               {"type": "text", "fields": {"keyword": { "type": "keyword"} } },
      "username":            {"type": "text", "fields": {"keyword": { "type": "keyword"} } },
      "password":            {"type": "keyword"},
      "description":         {"type": "text", "fields": {"keyword": { "type": "keyword"} } },
      "comment":             {"type": "text", "fields": {"keyword": { "type": "keyword"} } },
      "reg_date":            {"type": "date"},
      "location":            {"type": "geo_point"},
      "social_network_type": {"type": "keyword"}
    }
  }
}`

type Elastic struct {
	Client *elasticsearch.Client
	logger *slog.Logger
}

type elasticHit struct {
	ID     string      `json:"_id"`
	Source domain.User `json:"_source"`
}

type elasticResponse struct {
	Hits struct {
		Hits []elasticHit `json:"hits"`
	} `json:"hits"`
}

type elasticResponse2 struct {
    Index       string         `json:"_index"`
    ID          string         `json:"_id"`
    Version     int            `json:"_version"`
    SeqNo       int            `json:"_seq_no"`
    PrimaryTerm int            `json:"_primary_term"`
    Found       bool           `json:"found"`
    Source      domain.User    `json:"_source"`
}

var _ domain.UserRepository = (*Elastic)(nil) //проверка, что Elastic реализует интерфейс UserRepository

func Init(ctx context.Context, url string, logger *slog.Logger) (*Elastic, error) {
	const op = "elastic.Init"
	log := logger.With("operation", op)
	log.Debug("initializing Elasticsearch client", "url", url)

	start := time.Now()
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Error("client creation failed", "error", err)
		return nil, service.NewServiceError(
			service.ErrCodeInternal,
		)
	}
	log.Debug("client created", "duration", time.Since(start))
	res, err := client.Info()
	if err != nil {
		log.Error("connection test failed", "error", err)
		return nil, service.NewServiceError(
			service.ErrCodeInternal)
	}
	defer res.Body.Close()
	log.Debug("cluster info", "status", res.Status())
	exists, err := client.Indices.Exists([]string{usersIndex})
	if err != nil {
		log.Error("index check failed", "error", err)
		return nil, service.NewServiceError(
			service.ErrCodeInternal)
	}
	defer exists.Body.Close()

	if exists.StatusCode == 404 {
		log.Info("index not found, creating", "index", usersIndex)
		mapping := mappings

		createRes, err := client.Indices.Create(
			usersIndex,
			client.Indices.Create.WithBody(strings.NewReader(mapping)),
		)
		if err != nil {
			log.Error("index creation failed", "error", err)
			return nil, service.NewServiceError(
				service.ErrCodeInternal)
		}
		defer createRes.Body.Close()

		if createRes.IsError() {
			errBody, _ := io.ReadAll(createRes.Body)
			log.Error("index creation error", "response", string(errBody))
			return nil, service.NewServiceError(
				service.ErrCodeInternal)
		}
		log.Info("index created", "index", usersIndex)
	} else {
		log.Debug("index exists", "index", usersIndex)
	}

	log.Info("Elasticsearch initialized")
	return &Elastic{Client: client, logger: logger}, nil
}

func (e *Elastic) Create(ctx context.Context, user *domain.User) error {
	const op = "Elastic.Create"
	log := e.logger.With("operation", op, "user_id", user.ID)

	log.DebugContext(ctx, "creating user document")
	start := time.Now()

	if user.ID == nil || *user.ID == "" {
		uuid := uuid.New().String()
		user.ID = &uuid
		log.DebugContext(ctx, "generated new user ID")
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(user); err != nil {
		log.ErrorContext(ctx, "document encoding failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}

	res, err := e.Client.Create(
		usersIndex,
		*user.ID,
		&buf,
		e.Client.Create.WithContext(ctx),
		e.Client.Create.WithRefresh("wait_for"),
	)
	if err != nil {
		log.ErrorContext(ctx, "index request failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 409 {
			return service.NewServiceError(service.ErrCodeAlreadyExists)
		}
		log.ErrorContext(ctx, "index response error", "status", res.Status(), "response", res.String())
		return service.NewServiceError(service.ErrCodeInternal)
	}

	log.InfoContext(ctx, "user created", "duration", time.Since(start), "index", usersIndex)
	return nil
}

func (e *Elastic) GetByID(ctx context.Context, id *string) (*domain.User, error) {
	const op = "Elastic.GetByID"
	log := e.logger.With("operation", op, "user_id", id)
	log.DebugContext(ctx, "fetching user")
	start := time.Now()

	res, err := e.Client.Get(usersIndex, *id, e.Client.Get.WithContext(ctx))
	if err != nil {
		log.ErrorContext(ctx, "get request failed", "error", err)
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, service.NewServiceError(service.ErrCodeNotFound)
	}

	if res.IsError() {
		log.ErrorContext(ctx, "get response error", "status", res.Status(), "response", res.String())
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}

	var user *domain.User
	var elasticUser elasticResponse2
	if err := json.NewDecoder(res.Body).Decode(&elasticUser); err != nil {
		log.ErrorContext(ctx, "document decoding failed", "error", err)
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}
	user = &elasticUser.Source
	log.DebugContext(ctx, "user fetched", "duration", time.Since(start))
	return user, nil
}

func (e *Elastic) UpdatePartial(ctx context.Context, user *domain.User) error {
	const op = "Elastic.UpdatePartial"
	id := *user.ID
	user.ID = nil
	log := e.logger.With("operation", op, "user_id", id)
	log.DebugContext(ctx, "updating user", "fields", user)
	start := time.Now()

	updateBody := struct {
        Doc *domain.User `json:"doc"`
    }{
        Doc: user,
    }

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(updateBody); err != nil {
		log.ErrorContext(ctx, "document encoding failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}

	res, err := e.Client.Update(
		usersIndex,
		id,
		&buf,
		e.Client.Update.WithContext(ctx),
		e.Client.Update.WithRefresh("wait_for"),
	)
	if err != nil {
		log.ErrorContext(ctx, "update request failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return service.NewServiceError(service.ErrCodeNotFound)
	}

	if res.IsError() {
		log.ErrorContext(ctx, "update response error", "status", res.Status(), "response", res.String())
		return service.NewServiceError(service.ErrCodeInternal)
	}

	log.InfoContext(ctx, "user updated", "duration", time.Since(start))
	return nil
}

func (e *Elastic) Replace(ctx context.Context, user *domain.User) error {
	const op = "Elastic.Replace"
	log := e.logger.With("operation", op, "user_id", user.ID)

	log.DebugContext(ctx, "replace user", "fields", user)
	start := time.Now()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(user); err != nil {
		log.ErrorContext(ctx, "document encoding failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}

	res, err := e.Client.Index(
		usersIndex,
		&buf,
		e.Client.Index.WithDocumentID(*user.ID),
		e.Client.Index.WithContext(ctx),
		e.Client.Index.WithRefresh("wait_for"),
	)

	if err != nil {
		log.ErrorContext(ctx, "replace request failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		log.WarnContext(ctx, "user not found for replace")
		return service.NewServiceError(service.ErrCodeNotFound)
	}

	if res.IsError() {
		log.ErrorContext(ctx, "replace response error",
			"status", res.Status(),
			"response", res.String())
		return service.NewServiceError(service.ErrCodeInternal)
	}

	log.InfoContext(ctx, "user replaced",
		"duration", time.Since(start))
	return nil
}

func (e *Elastic) Delete(ctx context.Context, id *string) error {
	const op = "Elastic.Delete"
	log := e.logger.With("operation", op, "user_id", id)

	log.InfoContext(ctx, "deleting user")
	start := time.Now()

	res, err := e.Client.Delete(
		usersIndex,
		*id,
		e.Client.Delete.WithContext(ctx),
		e.Client.Delete.WithRefresh("wait_for"),
	)
	if err != nil {
		log.ErrorContext(ctx, "delete request failed", "error", err)
		return service.NewServiceError(service.ErrCodeInternal)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		log.WarnContext(ctx, "user not found for deletion")
		return service.NewServiceError(service.ErrCodeNotFound)
	}

	if res.IsError() {
		log.ErrorContext(ctx, "delete response error",
			"status", res.Status(),
			"response", res.String())
		return service.NewServiceError(service.ErrCodeInternal)
	}

	log.InfoContext(ctx, "user deleted",
		"duration", time.Since(start))
	return nil
}

func (e *Elastic) Search(ctx context.Context, filters *domain.UserFilter) ([]*domain.User, error) {
	const op = "Elastic.Search"
	log := e.logger.With("operation", op)

	log.DebugContext(ctx, "searching users", "filters", (*filters).String())
	start := time.Now()

	reader, err := buildElasticsearchQuery(filters)
	if err != nil {
		log.ErrorContext(ctx, "failed to build query", "error", err)
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}

	res, err := e.Client.Search(
		e.Client.Search.WithIndex(usersIndex),
		e.Client.Search.WithContext(ctx),
		e.Client.Search.WithBody(reader),
	)
	if err != nil {
		log.ErrorContext(ctx, "search request failed", "error", err)
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		log.WarnContext(ctx, "user not found for the search")
		return nil, service.NewServiceError(service.ErrCodeNotFound)
	}
	if res.IsError() {
		log.ErrorContext(ctx, "search response error",
			"status", res.Status(),
			"response", res.String())
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}

	var results []*domain.User
	if results, err = parseResults(res.Body); err != nil {
		log.ErrorContext(ctx, "document decoding failed", "error", err)
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}
	log.InfoContext(ctx, "search completed",
		"result_count", len(results),
		"duration", time.Since(start))

	return results, nil
}

func parseResults(r io.ReadCloser) ([]*domain.User, error) {

	var response elasticResponse
	if err := json.NewDecoder(r).Decode(&response); err != nil {
		return nil, service.NewServiceError(service.ErrCodeInternal)
	}

	return shoveTheId(response.Hits.Hits), nil
}
func (e *Elastic) Close() error {
	const op = "elastic.Close"
	e.logger.Debug("closing Elasticsearch client", "op", op)

	if transport, ok := e.Client.Transport.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
		e.logger.Debug("closed idle connections", "op", op)
	}
	e.logger.Info("elasticsearch client closed", "op", op)
	return nil
}

func shoveTheId(hits []elasticHit) []*domain.User {
	users := make([]*domain.User, len(hits))
	for i := range hits {
		user := &hits[i].Source

		id := hits[i].ID
		user.ID = &id

		users[i] = user
	}
	return users
}

func buildElasticsearchQuery(f *domain.UserFilter) (io.Reader, error) {
	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{},
			},
		},
	}

	mustQueries := []map[string]any{}
	if f.Search != nil && *f.Search != "" {
		searchQuery := map[string]any{
			"multi_match": map[string]any{
				"query": *f.Search,
				"fields": []string{
					"username",
					"login",
					"comment",
					"description",
				},
				"type": "best_fields",
			},
		}
		mustQueries = append(mustQueries, searchQuery)
	}
	if f.DateFrom != nil || f.DateTo != nil {
		rangeFilter := map[string]any{
			"range": map[string]any{
				"registration_date": map[string]any{},
			},
		}

		if f.DateFrom != nil {
			rangeFilter["range"].(map[string]any)["registration_date"].(map[string]any)["gte"] = f.DateFrom.Format(time.RFC3339)
		}
		if f.DateTo != nil {
			rangeFilter["range"].(map[string]any)["registration_date"].(map[string]any)["lte"] = f.DateTo.Format(time.RFC3339)
		}

		mustQueries = append(mustQueries, rangeFilter)
	}
	if f.Lat != nil && f.Lon != nil && f.Distance != nil {
		locationFilter := map[string]any{
			"geo_distance": map[string]any{
				"distance": *f.Distance,
				"location": map[string]any{
					"lat": f.Lat,
					"lon": f.Lon,
				},
			},
		}
		mustQueries = append(mustQueries, locationFilter)
	}
	if f.SocialType != nil && *f.SocialType != "" {
		socialTypeFilter := map[string]any{
			"term": map[string]any{
				"social_type": *f.SocialType,
			},
		}
		mustQueries = append(mustQueries, socialTypeFilter)
	}
	if len(mustQueries) == 0 {
		query["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	} else {
		query["query"].(map[string]any)["bool"].(map[string]any)["must"] = mustQueries
	}
	if f.SortBy != nil && *f.SortBy != "" {
		sortOrder := "asc"
		if f.SortOrder != nil && *f.SortOrder == "desc" {
			sortOrder = "desc"
		}

		sortField := *f.SortBy
		if sortField == "login" {
			sortField = "login.keyword"
		}

		query["sort"] = []map[string]any{
			{
				sortField: map[string]any{
					"order": sortOrder,
				},
			},
		}
	}
	from := 0
	size := 10

	if f.Size != nil && *f.Size > 0 {
		size = *f.Size
		if f.Page != nil && *f.Page > 0 {
			from = (*f.Page - 1) * size
		}
	}

	query["from"] = from
	query["size"] = size
	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
