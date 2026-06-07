package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"tralgo/types"

	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"goyave.dev/goyave/v5"
)

type ProviderHandler struct {
	Pool *pgxpool.Pool
}

func (h *ProviderHandler) ListProviders(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	rows, err := h.Pool.Query(ctx, "select provider_id, last_name, first_name, phone_number, email from providers")
	if err != nil {
		log.Println(err)
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		return
	}
	defer rows.Close()
	var providers []types.Provider
	for rows.Next() {
		var p types.Provider
		if err := rows.Scan(&p.Provider_id, &p.Last_name, &p.First_name, &p.Phone_number, &p.Email); err != nil {
			log.Println(err)
			response.Status(http.StatusInternalServerError)
			response.Error("Error reading the rows")
			return
		}
		log.Println(p.Provider_id, p.Last_name, p.First_name, p.Phone_number, p.Email)
		providers = append(providers, p)
	}
	response.JSON(http.StatusOK, providers)
}

func (h *ProviderHandler) CreateProvider(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	var provider types.Provider

	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&provider); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	// Data validation
	if provider.Last_name == "" || provider.First_name == "" {
		response.Status(http.StatusBadRequest)
		response.Error("Last name and First name required")
		return
	}
	if !strings.Contains(provider.Email, "@") || !strings.Contains(provider.Email, ".") {
		response.Status(http.StatusBadRequest)
		response.Error("Valid email adress is required")
		return
	}

	// Insert to db
	var providerID int
	err := h.Pool.QueryRow(ctx,
		`INSERT INTO providers (last_name, first_name, phone_number, email)
	VALUES ($1, $2, $3, $4)
	RETURNING provider_id`,
		provider.Last_name, provider.First_name, provider.Phone_number, provider.Email,
	).Scan(&providerID)

	if err != nil {
		response.Error(fmt.Sprintf("db query error: %s", err.Error()))
		return
	}
	provider.Provider_id = providerID
	response.JSON(http.StatusCreated, provider)
}

func (h *ProviderHandler) ShowProvider(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id_str := request.RouteParams["provider_id"]
	provider_id, err := strconv.Atoi(provider_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty provider_id")
		return
	}

	var p types.Provider
	err = h.Pool.QueryRow(ctx,
		"select provider_id, last_name, first_name, phone_number, email from providers where provider_id = $1",
		provider_id,
	).Scan(&p.Provider_id, &p.Last_name, &p.First_name, &p.Phone_number, &p.Email)

	if err == pgx.ErrNoRows {
		response.Status(http.StatusNotFound)
		response.Error("No provider found with this ID")
		return
	} else if err != nil {
		response.Error("db query error")
		return
	}
	response.JSON(http.StatusOK, p)
}

func (h *ProviderHandler) UpdateProvider(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id_str := request.RouteParams["provider_id"]
	provider_id, err := strconv.Atoi(provider_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty provider_id")
		return
	}

	var provider types.Provider
	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&provider); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}
	// Data validation
	if provider.Last_name == "" || provider.First_name == "" {
		response.Status(http.StatusBadRequest)
		response.Error("Last name and First name required")
		return
	}
	if !strings.Contains(provider.Email, "@") || !strings.Contains(provider.Email, ".") {
		response.Status(http.StatusBadRequest)
		response.Error("Valid email adress is required")
		return
	}

	tag, err := h.Pool.Exec(ctx,
		`UPDATE providers SET last_name = $1, first_name = $2, phone_number = $3, email = $4
     WHERE provider_id = $5`,
		provider.Last_name, provider.First_name, provider.Phone_number, provider.Email,
		provider_id,
	)
	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		log.Println(err)
		return
	}
	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No provider found for this id")
		return
	}
	provider.Provider_id = provider_id
	response.JSON(http.StatusOK, provider)
}

func (h *ProviderHandler) DeleteProvider(response *goyave.Response, request *goyave.Request) {

	ctx := request.Context()
	provider_id_str := request.RouteParams["provider_id"]
	provider_id, err := strconv.Atoi(provider_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty provider_id")
		return
	}

	tag, err := h.Pool.Exec(ctx,
		"delete from providers where provider_id = $1",
		provider_id,
	)
	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("Query to delete failed")
		return
	}

	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No provider found with this id")
		return
	}
	response.JSON(http.StatusOK, map[string]string{"deleted provider id": provider_id_str})
}
