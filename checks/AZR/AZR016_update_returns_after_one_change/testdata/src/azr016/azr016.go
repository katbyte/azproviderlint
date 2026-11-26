package azr016

import (
	"errors"
	"fmt"

	"azr016/schema"
)

type metadata struct{ ResourceData *schema.ResourceData }

func send(what string) error { return nil }

func read(d *schema.ResourceData) error { return nil }

func update(d *schema.ResourceData) error {
	if d.HasChange("sku_name") {
		if err := send("sku"); err != nil {
			return err
		}

		return read(d) // want `this returns once "sku_name" is handled, so a change to "tags" in the same apply is never sent`
	}

	if d.HasChange("tags") {
		if err := send("tags"); err != nil {
			return err
		}
	}

	return read(d)
}

func updateTyped(m metadata, key string) error {
	if m.ResourceData.HasChanges("zones", "capacity") {
		return nil // want `this returns once "zones" is handled, so a change to the properties below in the same apply is never sent`
	}

	if m.ResourceData.HasChange(key) {
		return send(key) // want `this returns once this change is handled, so a change to "tags" in the same apply is never sent`
	}

	payload := ""
	if m.ResourceData.HasChange("tags") {
		payload = "tags"
	}

	return send(payload)
}

// refusing a change is not finishing it, and the last change handled has nothing after it to miss
func fine(d *schema.ResourceData) error {
	if d.HasChange("name") {
		return errors.New("`name` cannot be changed")
	}

	if d.HasChange("location") {
		return fmt.Errorf("`location` cannot be changed to %s", "elsewhere")
	}

	if d.HasChange("sku_name") {
		if err := send("sku"); err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		return send("tags")
	}

	return nil
}

// an error from the work is passed on, and what is not about a change is not this rule's
func other(d *schema.ResourceData, ready bool) error {
	if d.HasChange("sku_name") {
		err := send("sku")
		return err
	}

	if ready {
		return nil
	}

	if d.HasChange("tags") {
		return send("tags")
	}

	return nil
}
