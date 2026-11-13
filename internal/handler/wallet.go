package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/kingperseocs/golazo-api/internal/domain"
)

// WalletDepositor es lo que el hanlder necesita del servicio del billeteras.
type WalletDepositor interface {
	Deposit(ctx context.Context, userID, walletID string, amountCents int64, ky string) (domain.WalletTrasaction, bool, error)
}

// WalletHanlder atiende las rutas HTTP de bileteras.
type WalletHandler struct {
	svc    WalletDepositor
	logger *slog.Logger
}

// NewWalletHanldler crea el handler con sus dependencias.
func NewWalletHandler(svc WalletDepositor, logger *slog.Logger) *WalletHandler {
	return &WalletHandler{svc: svc, logger: logger}
}

type depositRequest struct {
	AmountCents int64 `json:"amount_cents"`
}

// respuesta de la transacció
type transactionResponse struct {
	ID                string    `json:"id"`
	WalletID          string    `json:"wallet_id"`
	Type              string    `json:"type"`
	AmountCents       int64     `json:"amount_cents"`
	BalanceAfterCents int64     `json:"balance_after_cents"`
	CreatedAt         time.Time `json:"created_At"`
}

// Deposit atiende Post /wallets/{id}/deposits
func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		// solo pasa alguien registróla la ruta sin el middleware
		h.logger.Error("depósito sin usuario autenticado en el contexto")
		writeError(w, h.logger, http.StatusInternalServerError, "internal_error","ocurrió un error inesperado, intenta más tarde")
		return
	}

	walletID := r.PathValue("id")
	key := r.Header.Get("Idempotency-Key")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req depositRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.logger.Warn("JSON inválido en depósito", "error", err)
		writeError(w, h.logger, http.StatusBadRequest, "invalid_json", "el cuerpo debe ser un JSON válido")
		return
	}

	tx, created, err := h.svc.Deposit(r.Context(), userID, walletID, req.AmountCents, key)
	if err != nil {
		h.writeDepositError(w, err)
		return
	}

	status := http.StatusCreated
	if !created {
		status = http.StatusOK // reintento: mismo resultado, sin acreditar de nuevo
	}

	writeJSON(w, h.logger, status, transactionResponse{
		ID:                tx.ID,
		WalletID:          tx.WalletID,
		Type:              string(tx.Type),
		AmountCents:       tx.AmountCents,
		BalanceAfterCents: tx.BalanceAfterCents,
		CreatedAt:         tx.CreateAt,
	})

}

// writeDepositError traduce los erores del depósito a respuestas HTTP.
func (h *WalletHandler) writeDepositError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrWalletNotFound):
		writeError(w, h.logger, http.StatusNotFound, "wallet_not_found", err.Error())
	case errors.Is(err, domain.ErrIdempotencyConflict):
		writeError(w, h.logger, http.StatusConflict, "idempotency_conflict", err.Error())
	case errors.Is(err, domain.ErrInvalidAmount):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_amount", err.Error())
	case errors.Is(err, domain.ErrDepositLimit):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "deposit_limit", err.Error())
	case errors.Is(err, domain.ErrInvalidIdempotency):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_idempotency_key", err.Error())
	default:
		h.logger.Error("depósito falló", "error", err)
		writeError(w, h.logger, http.StatusInternalServerError, "internal_error", "ocurrió un error interno, intenta más tarde")
	}
}
