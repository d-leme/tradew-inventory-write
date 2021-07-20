package inventory

import (
	"context"
	"strings"
	"time"

	"github.com/Tra-Dew/inventory-write/pkg/core"
)

// ItemStatus ...
type ItemStatus string

const (
	// ItemAvailable is set when an item is available to be traded
	ItemAvailable ItemStatus = "Available"
)

// ItemName ...
type ItemName string

// ItemDescription ...
type ItemDescription string

// ItemQuantity ...
type ItemQuantity int64

// Item ...
type Item struct {
	ID             string
	OwnerID        string
	Name           ItemName
	Status         ItemStatus
	Description    *ItemDescription
	TotalQuantity  ItemQuantity
	LockedQuantity ItemQuantity
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Repository ...
type Repository interface {
	InsertBulk(ctx context.Context, items []*Item) error
	UpdateBulk(ctx context.Context, userID string, items []*Item) error
	DeleteBulk(ctx context.Context, userID string, ids []string) error
	Get(ctx context.Context, userID string, ids []string) ([]*Item, error)
}

// Service ...
type Service interface {
	CreateItems(ctx context.Context, userID, correlationID string, req *CreateItemsRequest) error
	UpdateItems(ctx context.Context, userID, correlationID string, req *UpdateItemsRequest) error
	LockItems(ctx context.Context, userID, correlationID string, req *LockItemsRequest) error
	DeleteItems(ctx context.Context, userID, correlationID string, req *DeleteItemsRequest) error
}

// NewItemName ...
func NewItemName(name string) (ItemName, error) {
	name = strings.TrimSpace(name)
	if len(name) < 3 {
		return "", core.ErrValidationFailed
	}

	return ItemName(name), nil
}

// NewItemDescription ...
func NewItemDescription(description *string) *ItemDescription {
	if description == nil || *description == "" {
		return nil
	}

	itemDescription := ItemDescription(strings.TrimSpace(*description))

	return &itemDescription
}

// NewItemQuantity ...
func NewItemQuantity(quantity int64) (ItemQuantity, error) {
	if quantity <= 0 {
		return 0, core.ErrValidationFailed
	}

	return ItemQuantity(quantity), nil
}

// NewItem ...
func NewItem(id, ownerID, name string, description *string, quantity int64, status ItemStatus) (*Item, error) {

	if id == "" {
		return nil, core.ErrValidationFailed
	}

	if ownerID == "" {
		return nil, core.ErrValidationFailed
	}

	itemName, err := NewItemName(name)
	if err != nil {
		return nil, err
	}

	itemQuantity, err := NewItemQuantity(quantity)
	if err != nil {
		return nil, err
	}

	if status == "" {
		return nil, core.ErrValidationFailed
	}

	return &Item{
		ID:             id,
		OwnerID:        ownerID,
		Name:           itemName,
		Status:         status,
		Description:    NewItemDescription(description),
		TotalQuantity:  itemQuantity,
		LockedQuantity: 0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// Update ...
func (item *Item) Update(name string, description *string, quantity int64) error {
	itemName, err := NewItemName(name)
	if err != nil {
		return err
	}

	itemDescription := NewItemDescription(description)
	if err != nil {
		return err
	}

	itemQuantity, err := NewItemQuantity(quantity)
	if err != nil {
		return err
	}

	item.Name = itemName
	item.Description = itemDescription
	item.TotalQuantity = itemQuantity
	item.UpdatedAt = time.Now()

	return nil
}

// Lock ...
func (item *Item) Lock(quantity int64) error {

	itemQuantity, err := NewItemQuantity(quantity)
	if err != nil {
		return err
	}

	lockedQuantity := item.LockedQuantity + itemQuantity

	if lockedQuantity > item.TotalQuantity {
		return core.ErrNotEnoughtItemsToLock
	}

	item.LockedQuantity = lockedQuantity
	item.UpdatedAt = time.Now()

	return nil
}
