package example

type Status string

func PossibleValuesForStatus() []string {
	return []string{"Enabled", "Disabled"}
}

type StatusAlias = Status

type ResourceIdentifier string

type Count int

func PossibleValuesForCount() []int {
	return []int{1, 2}
}

// Helpers with the wrong signature below do not identify an SDK enum

type WrongParams string

func PossibleValuesForWrongParams(string) []string {
	return nil
}

type WrongResult string

func PossibleValuesForWrongResult() string {
	return ""
}

type WrongElement string

func PossibleValuesForWrongElement() []int {
	return nil
}

type NotFunction string

var PossibleValuesForNotFunction = 1
