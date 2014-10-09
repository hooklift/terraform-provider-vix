// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vmx

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Decoder struct {
	// Scanner to read file line by line
	scanner *bufio.Scanner

	// This map was basically created to reduce runtime complexity from O(n^2)
	// to O(n). The trade-off is that more memory will be used which seems to be
	// a fair trade-off.
	vmx map[string]string

	// Report an error if there are keys in the Go structure
	// that do not have a match in the VMX file
	ErrorUnmatched bool
}

func NewDecoder(reader io.Reader, errorUnmatched bool) *Decoder {
	return &Decoder{
		scanner:        bufio.NewScanner(reader),
		ErrorUnmatched: errorUnmatched,
	}
}

// Loads VMX file in a map so we can do searches in O(1)
func (d *Decoder) loadVMXMap() error {
	errors := make([]string, 0)

	if len(d.vmx) == 0 {
		d.vmx = make(map[string]string)
	}

	for d.scanner.Scan() {
		line := d.scanner.Text()

		// Ignore comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			errors = appendErrors(errors, fmt.Errorf("Invalid line: %s ", line))
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		unquotedVal, err := strconv.Unquote(value)
		if err != nil {
			unquotedVal = value
			//errors = appendErrors(errors, fmt.Errorf("Error unquoting vmx value %s: %v.", value, err))
		}

		key = strings.ToLower(key)
		d.vmx[key] = unquotedVal
	}

	if err := d.scanner.Err(); err != nil {
		return fmt.Errorf("Scanner error: %v", err)
	}

	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

// Decodes a VMX file into a Go structure. This is O(n) for now, we can
// optimize later if needed. Although, runtime complexities will vary depending
// on the value type and whether there are nested fields in the reflect value.
// However, the impact could be minimized by setting bounds to the recursion.
func (d *Decoder) Decode(v interface{}) error {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer value passed to Decode")
	}

	if val.IsNil() {
		return fmt.Errorf("nil value passed to Decode: %v", reflect.TypeOf(val))
	}

	// Gets setteable value
	val = val.Elem()

	if !val.CanAddr() {
		return errors.New("destination struct must be addressable")
	}

	err := d.loadVMXMap()
	if err != nil {
		return err
	}

	return d.decode(val, "")
}

// Lets decode only what the reflect value is asking for as opposed to starting
// by iterating the text file. This strategy makes it exponentially
// easier to bind to slices or arrays. For the future myself looking at this code,
// trust me, I tried first the other strategy and got stuck on that. For other
// contributors, I'm happy to discuss more if you ask me.
// -- c4milo
func (d *Decoder) decode(val reflect.Value, key string) error {
	//fmt.Printf("[D] Decoding into key ->%s<-...\n", key)
	errors := make([]string, 0)

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := string(typeField.Tag)
		valueField := val.Field(i)

		// Internal govmx way of identifying elements in a slice so that
		// users can delete them individually.
		if typeField.Name == "VMXID" {
			valueField.SetString(key)
			continue
		}

		// Ignore untagged fields only if they are not embedded structs
		if tag == "" && !typeField.Anonymous {
			continue
		}

		destKey, _, err := parseTag(tag)
		if err != nil {
			continue
		}

		if key != "" {
			if destKey != "" {
				destKey = key + "." + destKey
			} else {
				destKey = key
			}
		}

		destKey = strings.ToLower(destKey)

		if destKey == "-" || !valueField.CanSet() {
			log.Printf("Cant set type %s tagged as %s\n", valueField.Type().String(), destKey)
			continue
		}

		err = d.reflectKind(valueField.Kind(), valueField, destKey)
		if err != nil {
			errors = appendErrors(errors, err)
		}
	}

	if len(errors) > 0 {
		return &Error{errors}
	}
	return nil
}

func (d *Decoder) reflectKind(kind reflect.Kind, valueField reflect.Value, key string) error {
	var err error

	value := d.vmx[key]
	//fmt.Printf("%s => %s\n", key, value)

	if kind != reflect.Struct && kind != reflect.Array &&
		kind != reflect.Slice && kind != reflect.Map {
		if value == "" {
			if d.ErrorUnmatched {
				return fmt.Errorf("Unmatched key found in Go type: %s", key)
			}
			return nil
		}
	}

	switch kind {
	case reflect.Struct:
		err = d.decode(valueField, key)

	case reflect.Array, reflect.Slice:
		err = d.decodeSlice(valueField, key)

	case reflect.Map:
		// TODO(c4milo)
	case reflect.String:
		valueField.SetString(value)

	case reflect.Bool:
		var boolValue bool
		boolValue, err = strconv.ParseBool(value)
		valueField.SetBool(boolValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intValue int64
		intValue, err = strconv.ParseInt(value, 10, valueField.Type().Bits())
		valueField.SetInt(intValue)

	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		var uintValue uint64
		uintValue, err = strconv.ParseUint(value, 10, valueField.Type().Bits())
		valueField.SetUint(uintValue)

	default:
		err = fmt.Errorf("Data type unsupported: %s", valueField.Kind())
	}

	return err
}

func (d *Decoder) decodeSlice(valueField reflect.Value, key string) error {
	//fmt.Printf("[D] Decode slice tagged as: ->%s<-\n", key)

	errors := make([]string, 0)
	seenIndexes := make(map[string]bool)

	for k, _ := range d.vmx {
		if !strings.HasPrefix(k, key) {
			continue
		}

		index := getVMXAttrIndex(k, key)
		if index == "" || seenIndexes[index] {
			continue
		}

		// The reason we have to keep track of seen indexes is because entries
		// in the vmx file with the same prefix are actually objects, they are
		// decoded into Go structs, meaning that they only need one pass to be decoded.
		seenIndexes[index] = true

		length := valueField.Len()
		capacity := valueField.Cap()

		// Grow the slice if needed. This allows us to pass a value
		// reference to d.decode() so it populates the value addressed by the slice.
		if length >= capacity {
			capacity := 2 * length
			if capacity < 4 {
				capacity = 4
			}

			newSlice := reflect.MakeSlice(valueField.Type(), length, capacity)
			reflect.Copy(newSlice, valueField)
			valueField.Set(newSlice)
		}

		valueField.SetLen(length + 1)

		newKey := key
		if key != index {
			newKey = key + index
		}

		err := d.decode(valueField.Index(length), newKey)

		if err != nil {
			errors = appendErrors(errors, err)
			valueField.SetLen(length)
		}
	}

	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

// In a VMX entry, an index is the prefix, before the first dot, minus the
// attribute involved as declared in the given Go tag by the user.
//
// Examples:
//   - In ethernet1.addressType the index is 1.
//   - In scsi0:0.filename the index is 0:0
//   - In usb:1.deviceType the index is :1
//   - In ide1:0.filename the index is 1:0
func getVMXAttrIndex(vmxKey, key string) string {
	// trimming the attribute's name returns 1.present in the case of ethernet1.present,
	// 0:0.filename for scsi0:0.filename, or :1.present for usb:1.present
	attr := strings.TrimPrefix(vmxKey, key)

	parts := strings.Split(attr, ".")

	index := ""
	if len(parts) > 0 {
		index = parts[0]
	}

	if index == "" {
		return key
	}

	// If it is a disk controller, get the controller's index
	// parts2 := strings.Split(index, ":")
	// if len(parts2) > 0 {
	// 	index = parts2[0]
	// }

	return index
}
