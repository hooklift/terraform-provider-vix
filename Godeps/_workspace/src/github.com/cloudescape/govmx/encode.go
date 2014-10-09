// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vmx

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type Encoder struct {
	// Buffer where the vmx file will be written to
	buffer *bytes.Buffer
	// Maximum recursion allowed
	maxRecursion uint8
	// Current recursion level
	currentRecursion uint8
	// Parent key, used for recursion. This will allow us to set the correct
	// keys for nested structures.
	parentKey string
}

// Creates a new encoder
func NewEncoder(buffer *bytes.Buffer) *Encoder {
	return &Encoder{
		buffer:       buffer,
		maxRecursion: 5,
	}
}

// Encodes Go structure into a VMX structure, recursively.
func (e *Encoder) Encode(v interface{}) error {
	val := reflect.ValueOf(v)
	return e.encode(val)
}

// Does the actual encoding work
func (e *Encoder) encode(val reflect.Value) error {
	// Drill into interfaces and pointers.
	// This can turn into an infinite loop given a cyclic chain,
	// but it matches the Go 1 behavior.
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		tag := typeField.Tag

		key, omitempty, err := parseTag(string(tag))
		if err != nil {
			return err
		}

		if key == "-" || !valueField.IsValid() ||
			omitempty && isEmptyValue(valueField) ||
			key == "" && !typeField.Anonymous {
			continue
		}

		kind := valueField.Kind()
		switch kind {
		case reflect.Struct:
			err = e.encodeStruct(valueField, typeField, key)
		case reflect.Array, reflect.Slice:
			err = e.encodeArray(valueField, key)
		default:
			if e.parentKey != "" {
				key = e.parentKey + "." + key
			}

			//fmt.Printf("parent key: %s, key: %s \n", e.parentKey, key)
			value := valueField.Interface()
			e.buffer.WriteString(fmt.Sprintf("%s = \"%v\"\n", key, value))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// When an array or slice type is found in the Go structure, this function encodes it
// recursively.
func (e *Encoder) encodeArray(valueField reflect.Value, key string) error {
	adaptersCnt := 0
	devicesCnt := 0

	for i := 0; i < valueField.Len(); i++ {
		switch key {
		case "ide":
			if i >= (MAX_IDE_ADAPTERS * MAX_IDE_DEVICES_PER_ADAPTER) {
				return nil
			}

			if devicesCnt >= MAX_IDE_DEVICES_PER_ADAPTER {
				adaptersCnt++
				devicesCnt = 0
			}

			e.parentKey = fmt.Sprintf("%s%d:%d", key, adaptersCnt, devicesCnt)
			devicesCnt++

		case "scsi":
			if i >= (MAX_SCSI_ADAPTERS * MAX_SCSI_DEVICES_PER_ADAPTER) {
				return nil
			}

			if devicesCnt >= MAX_SCSI_DEVICES_PER_ADAPTER {
				adaptersCnt++
				devicesCnt = 0
			}

			val := valueField.Index(i).FieldByName("VirtualDev")
			if !isEmptyValue(val) {
				e.parentKey = fmt.Sprintf("%s%d", key, adaptersCnt)
			} else {
				e.parentKey = fmt.Sprintf("%s%d:%d", key, adaptersCnt, devicesCnt)
				devicesCnt++
			}
		case "sata":
			if i >= (MAX_SATA_ADAPTERS * MAX_SATA_DEVICES_PER_ADAPTER) {
				return nil
			}

			if devicesCnt >= MAX_SATA_DEVICES_PER_ADAPTER {
				adaptersCnt++
				devicesCnt = 0
			}

			e.parentKey = fmt.Sprintf("%s%d:%d", key, adaptersCnt, devicesCnt)
			devicesCnt++
		case "usb":
			if i >= MAX_USB_ADAPTERS*MAX_USB_DEVICES {
				return nil
			}

			e.parentKey = fmt.Sprintf("%s:%d", key, devicesCnt)
			devicesCnt++
		case "ethernet":
			if i >= MAX_VNICS {
				return nil
			}
			e.parentKey = key + strconv.Itoa(i)
		default:
			e.parentKey = key + strconv.Itoa(i)
		}

		err := e.encode(valueField.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// Encodes a Go struct type into a VMX string, recursively.
func (e *Encoder) encodeStruct(valueField reflect.Value, typeField reflect.StructField, key string) error {
	e.currentRecursion++
	if e.currentRecursion > e.maxRecursion {
		return nil
	}

	e.parentKey += key
	err := e.encode(valueField)
	if err != nil {
		return err
	}

	// Do not reset the key if we are dealing with an embedded struct
	// as recursion level increases by one compared to normal cases
	if !typeField.Anonymous {
		e.parentKey = ""
	}

	e.currentRecursion--
	return nil
}

// Checks whether or not the reflected value is empty
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
