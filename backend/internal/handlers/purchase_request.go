package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"starter-backend/internal/models"
)

var (
	errPurchaseRequestConflict = errors.New("purchase request key reused with different payload")
	errPurchasePriceChanged    = errors.New("purchase total differs from client expectation")
	purchaseRequestKeyPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
)

type purchaseRequestRecord struct {
	TouristID   uint64
	RequestKey  string
	PayloadHash string
	PurchaseID  uint64
}

func purchasePayloadHash(input purchaseCreateInput) string {
	raw, _ := json.Marshal(input)
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}

// The transaction lock serializes identical keys before checking the ledger.
// Both purchase, capacity update and ledger entry commit or roll back together.
func replayPurchaseRequest(tx *gorm.DB, userID uint64, key, hash string, purchase *models.TourPackagePurchase) (bool, error) {
	if key == "" {
		return false, nil
	}
	lockKey := fmt.Sprintf("purchase:%d:%s", userID, key)
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockKey).Error; err != nil {
		return false, err
	}
	var record purchaseRequestRecord
	err := tx.Table("purchase_requests").Where("tourist_id = ? AND request_key = ?", userID, key).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if record.PayloadHash != hash {
		return false, errPurchaseRequestConflict
	}
	err = tx.Where("id = ? AND tourist_id = ?", record.PurchaseID, userID).First(purchase).Error
	return err == nil, err
}
