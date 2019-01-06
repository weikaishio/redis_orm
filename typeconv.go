package redis_orm

import (
	"encoding/json"
	"strconv"
)

func ToString(v interface{}) string {
	switch value := v.(type) {
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(int64(value), 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(uint64(value), 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64)
	case string:
		return value
	case []byte:
		return string(value)
	case bool:
		if value {
			return "true"
		} else {
			return "false"
		}
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "{}"
		}
		return string(data)
	}
}

func SetInt64FromStr(ptr *int64, s string) error {
	i, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		*ptr = i
	}
	return err
}

func SetInt32FromStr(ptr *int32, s string) error {
	i, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		*ptr = int32(i)
	}
	return err
}

func SetIntFromStr(ptr *int, s string) error {
	i, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		*ptr = int(i)
	}
	return err
}

func SetUint64FromStr(ptr *uint64, s string) error {
	i, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*ptr = i
	}
	return err
}

func SetUint32FromStr(ptr *uint32, s string) error {
	i, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*ptr = uint32(i)
	}
	return err
}

func SetUint16FromStr(ptr *uint16, s string) error {
	i, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*ptr = uint16(i)
	}
	return err
}

func SetUint8FromStr(ptr *uint8, s string) error {
	i, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*ptr = uint8(i)
	}
	return err
}

func SetUintFromStr(ptr *uint, s string) error {
	i, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*ptr = uint(i)
	}
	return err
}

func SetFloat32FromStr(ptr *float32, s string) error {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		*ptr = float32(f)
	}
	return err
}

func SetFloat64FromStr(ptr *float64, s string) error {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		*ptr = float64(f)
	}
	return err
}

func SetBoolFromStr(ptr *bool, s string) error {
	if s == "" {
		*ptr = false
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err == nil {
		*ptr = b
	}
	return err
}
