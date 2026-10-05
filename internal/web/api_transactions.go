package web

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtransactions"
)

type transactionRequest struct {
	AssetID   int64   `json:"assetId"`
	Kind      string  `json:"kind"`
	AmountUSD float64 `json:"amountUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}

type transactionUpdateRequest struct {
	Kind      string  `json:"kind"`
	AmountUSD float64 `json:"amountUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}

type transactionBatchRequest struct {
	AssetID int64                      `json:"assetId"`
	Items   []transactionUpdateRequest `json:"items"`
}

type transactionMonthRequest struct {
	Month int                           `json:"month"`
	Year  int                           `json:"year"`
	Items []transactionMonthRequestItem `json:"items"`
}

type transactionMonthRequestItem struct {
	AssetID   int64   `json:"assetId"`
	AmountUSD float64 `json:"amountUsd"`
}

type idsResponse struct {
	IDs []int64 `json:"ids"`
}

type assetTransactionsResponse struct {
	Asset  domain.Asset            `json:"asset"`
	Rows   []viewtransactions.Row  `json:"rows"`
	Totals viewtransactions.Totals `json:"totals"`
}

type monthAssetsResponse struct {
	Period   domain.YearMonth `json:"period"`
	Eligible []domain.Asset   `json:"eligible"`
	Omitted  []domain.Asset   `json:"omitted"`
}

// registerTransactionRoutes rexistra as rutas de transaccións da API.
func (s *Server) registerTransactionRoutes() {
	s.handleAPI("GET /api/assets/{id}/transactions", s.listAssetTransactions)
	s.handleAPI("POST /api/transactions", s.createTransaction)
	s.handleAPI("POST /api/transactions/batch", s.createTransactionBatch)
	s.handleAPI("GET /api/transactions/month-assets", s.listTransactionMonthAssets)
	s.handleAPI("POST /api/transactions/month", s.createTransactionMonth)
	s.handleAPI("PUT /api/transactions/{id}", s.updateTransaction)
	s.handleAPI("DELETE /api/transactions/{id}", s.deleteTransaction)
}

func (s *Server) listAssetTransactions(w http.ResponseWriter, r *http.Request) {
	asset, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	txs, err := s.store.ListTransactionsByAsset(asset.ID)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	rows := viewtransactions.BuildRows(asset, txs)
	writeJSON(w, http.StatusOK, assetTransactionsResponse{Asset: asset, Rows: nonNil(rows), Totals: viewtransactions.ComputeTotals(rows)})
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req transactionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	tx, ok, failed := s.transactionFromRequest(w, r, fields, "", req.AssetID, req.Kind, req.AmountUSD, req.Year, req.Month)
	if failed {
		return
	}
	if !ok {
		writeFieldErrors(w, fields)
		return
	}
	id, err := s.store.InsertTransaction(tx)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	tx.ID = id
	writeJSON(w, http.StatusCreated, tx)
}

func (s *Server) createTransactionBatch(w http.ResponseWriter, r *http.Request) {
	var req transactionBatchRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	asset, assetOK, failed := s.bodyAsset(w, r, fields, "assetId", req.AssetID)
	if failed {
		return
	}
	txs := make([]domain.Transaction, 0, len(req.Items))
	for i, item := range req.Items {
		prefix := fmt.Sprintf("items[%d].", i)
		validateKindAmountDate(fields, prefix, item.Kind, item.AmountUSD, item.Year, item.Month)
		if assetOK {
			validateDateNotBefore(fields, prefix+"month", domain.YearMonth{Year: item.Year, Month: item.Month}, asset.Start())
		}
		amount, _ := signedTransactionAmount(item.Kind, item.AmountUSD)
		txs = append(txs, domain.Transaction{AssetID: req.AssetID, AmountUSD: amount, Year: item.Year, Month: item.Month})
	}
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	ids, err := s.store.InsertTransactions(txs)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, idsResponse{IDs: nonNil(ids)})
}

func (s *Server) listTransactionMonthAssets(w http.ResponseWriter, r *http.Request) {
	period, fields := queryYearMonth(r)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	assets, err := s.store.ListAssets()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	var eligible, omitted []domain.Asset
	for _, asset := range assets {
		if asset.CreatedBy(period) {
			eligible = append(eligible, asset)
		} else {
			omitted = append(omitted, asset)
		}
	}
	writeJSON(w, http.StatusOK, monthAssetsResponse{Period: period, Eligible: nonNil(eligible), Omitted: nonNil(omitted)})
}

func (s *Server) createTransactionMonth(w http.ResponseWriter, r *http.Request) {
	var req transactionMonthRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	validateYearMonth(fields, "year", "month", req.Year, req.Month)
	period := domain.YearMonth{Year: req.Year, Month: req.Month}
	seen := map[int64]bool{}
	txs := make([]domain.Transaction, 0, len(req.Items))
	for i, item := range req.Items {
		prefix := fmt.Sprintf("items[%d].", i)
		validateAmount(fields, prefix+"amountUsd", item.AmountUSD)
		asset, ok, failed := s.bodyAsset(w, r, fields, prefix+"assetId", item.AssetID)
		if failed {
			return
		}
		if seen[item.AssetID] {
			fields[prefix+"assetId"] = "o ativo está repetido"
		}
		seen[item.AssetID] = true
		if ok && period.Valid() && !asset.CreatedBy(period) {
			fields[prefix+"assetId"] = "o ativo creouse despois dese mes"
		}
		txs = append(txs, domain.Transaction{AssetID: item.AssetID, AmountUSD: item.AmountUSD, Year: req.Year, Month: req.Month})
	}
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	ids, err := s.store.InsertTransactions(txs)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, idsResponse{IDs: nonNil(ids)})
}

func (s *Server) updateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	current, err := s.store.GetTransaction(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transacción non atopada")
			return
		}
		s.internalError(w, r, err)
		return
	}
	var req transactionUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	asset, assetOK, failed := s.bodyAsset(w, r, fields, "assetId", current.AssetID)
	if failed {
		return
	}
	validateKindAmountDate(fields, "", req.Kind, req.AmountUSD, req.Year, req.Month)
	if assetOK {
		validateDateNotBefore(fields, "month", domain.YearMonth{Year: req.Year, Month: req.Month}, asset.Start())
	}
	if len(fields) > 0 {
		delete(fields, "assetId")
		writeFieldErrors(w, fields)
		return
	}
	amount, _ := signedTransactionAmount(req.Kind, req.AmountUSD)
	updated := domain.Transaction{ID: id, AssetID: current.AssetID, AmountUSD: amount, Year: req.Year, Month: req.Month}
	if err := s.store.UpdateTransaction(updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transacción non atopada")
			return
		}
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.DeleteTransaction(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transacción non atopada")
			return
		}
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) transactionFromRequest(w http.ResponseWriter, r *http.Request, fields map[string]string, prefix string, assetID int64, kind string, amount float64, year, month int) (domain.Transaction, bool, bool) {
	asset, assetOK, failed := s.bodyAsset(w, r, fields, prefix+"assetId", assetID)
	if failed {
		return domain.Transaction{}, false, true
	}
	validateKindAmountDate(fields, prefix, kind, amount, year, month)
	if assetOK {
		validateDateNotBefore(fields, prefix+"month", domain.YearMonth{Year: year, Month: month}, asset.Start())
	}
	if len(fields) > 0 {
		return domain.Transaction{}, false, false
	}
	stored, _ := signedTransactionAmount(kind, amount)
	return domain.Transaction{AssetID: assetID, AmountUSD: stored, Year: year, Month: month}, true, false
}

func validateKindAmountDate(fields map[string]string, prefix, kind string, amount float64, year, month int) {
	if _, ok := signedTransactionAmount(kind, amount); !ok {
		fields[prefix+"kind"] = "o tipo de transacción debe ser compra ou venda"
	}
	validateAmount(fields, prefix+"amountUsd", amount)
	validateYearMonth(fields, prefix+"year", prefix+"month", year, month)
}
