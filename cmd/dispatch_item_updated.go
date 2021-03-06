package cmd

import (
	"context"

	"github.com/d-leme/tradew-inventory-write/pkg/core"
	"github.com/d-leme/tradew-inventory-write/pkg/inventory"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// DispatchItemUpdated ...
func DispatchItemUpdated(command *cobra.Command, args []string) {
	settings := new(core.Settings)

	err := core.FromYAML(command.Flag("settings").Value.String(), settings)
	if err != nil {
		logrus.
			WithError(err).
			Fatal("unable to parse settings, shutting down...")
	}

	ctx := context.Background()
	container := NewContainer(settings)

	items, err := container.InventoryRepository.GetByStatus(ctx, inventory.ItemPendingUpdateDispatch)
	if err != nil {
		logrus.WithError(err).Error("error while getting items by status")
		return
	}

	lenItems := len(items)

	logrus.Infof("%d new items to publish", lenItems)

	if lenItems < 1 {
		return
	}

	event := inventory.ParseItemsToItemsUpdatedEvent(items)
	messageID, err := container.Producer.Publish(settings.Events.ItemsUpdated, event)

	if err != nil {
		logrus.WithError(err).Error("error while dispatching message")
		return
	}

	fields := logrus.Fields{"message_id": messageID}

	logrus.WithFields(fields).Info("dipached event")

	for _, item := range items {
		item.UpdateStatus(inventory.ItemAvailable)
	}

	if err := container.InventoryRepository.UpdateBulk(ctx, items); err != nil {
		logrus.WithError(err).WithFields(fields).Error("error while updating items")
		return
	}

	logrus.
		WithFields(fields).
		Info("worker complete")
}
