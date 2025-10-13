package conversion

import "github.com/google/uuid"

func UUIDToString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case uuid.UUID:
		return v.String()
	default:
		return ""
	}
}

func ToUUID(val interface{}) (uuid.UUID, error) {
	if val == nil {
		return uuid.Nil, nil
	}
	switch v := val.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, nil
	}
}
