package utils

import (
	"fmt"
	"testing"
)

type jsonfrom struct {
	Element1 string `json:"element2"`
	Element2 int    `json:"element1"`
}
type jsonto struct {
	Elt1 int    `json:"element1"`
	Elt2 string `json:"element2"`
	Elt3 int
}

func TestInParallel(t *testing.T) {
	size := 5
	booleanMap := make(map[int]bool, size+1)
	tasks := InParallel{}
	for i := 0; i < size; i++ {
		task := func(index int) func() error {
			return func() error {
				booleanMap[index] = true
				return nil
			}
		}(i)
		tasks.Add(task)
	}
	err := tasks.Wait()
	if err != nil {
		t.Fatal(err)
	}
	// Make sure every value is true
	for _, value := range booleanMap {
		if !value {
			t.Fatalf("booleanMap expected to contain only true values, got at least one false")
		}
	}
}

func TestInParallelError(t *testing.T) {
	size := 5
	booleanMap := make(map[int]bool, size+1)
	tasks := InParallel{}
	for i := 0; i < size; i++ {
		task := func(index int) func() error {
			return func() error {
				booleanMap[index] = true
				if index%2 == 0 {
					return fmt.Errorf("Error with %v", index)
				}
				return nil
			}
		}(i)
		tasks.Add(task)
	}
	err := tasks.Wait()
	if err == nil {
		t.Fatalf("Expected an error on Wait, got nothing.")
	}
	for key, value := range booleanMap {
		if key%2 != 0 && !value {
			t.Fatalf("booleanMap expected to contain true values on odd number, got %v", booleanMap)
		}
	}
}

func TestConvertByJSON(t *testing.T) {
	valids := []struct {
		src      jsonfrom
		expected jsonto
	}{
		{
			jsonfrom{Element2: 1},
			jsonto{1, "", 0},
		},
		{
			jsonfrom{},
			jsonto{0, "", 0},
		},
		{
			jsonfrom{"element1", 2},
			jsonto{2, "element1", 0},
		},
	}
	for _, valid := range valids {
		var target jsonto
		err := ConvertByJSON(valid.src, &target)
		if err != nil || target.Elt1 != valid.expected.Elt1 || target.Elt2 != valid.expected.Elt2 || target.Elt3 != 0 {
			t.Fatalf("Expected %v from %v got %v, %v", valid.expected, valid.src, target, err)
		}
	}
}

func TestConvertByJSONInvalid(t *testing.T) {
	invalids := []interface{}{
		// Incompatible struct
		struct {
			Element1 int    `json:"element2"`
			Element2 string `json:"element1"`
		}{1, "element1"},
		// Not marshable struct
		struct {
			Element1 func(int) int
		}{
			func(i int) int { return 0 },
		},
	}
	for _, invalid := range invalids {
		var target jsonto
		if err := ConvertByJSON(invalid, &target); err == nil {
			t.Fatalf("Expected an error converting %v to %v, got nothing", invalid, target)
		}
	}
}

type yamlfrom struct {
	Element1 string `yaml:"element2"`
	Element2 int    `yaml:"element1"`
}
type yamlto struct {
	Elt1 int    `yaml:"element1"`
	Elt2 string `yaml:"element2"`
	Elt3 int
}

func TestConvert(t *testing.T) {
	valids := []struct {
		src      yamlfrom
		expected yamlto
	}{
		{
			yamlfrom{Element2: 1},
			yamlto{1, "", 0},
		},
		{
			yamlfrom{},
			yamlto{0, "", 0},
		},
		{
			yamlfrom{"element1", 2},
			yamlto{2, "element1", 0},
		},
	}
	for _, valid := range valids {
		var target yamlto
		err := Convert(valid.src, &target)
		if err != nil || target.Elt1 != valid.expected.Elt1 || target.Elt2 != valid.expected.Elt2 || target.Elt3 != 0 {
			t.Fatalf("Expected %v from %v got %v, %v", valid.expected, valid.src, target, err)
		}
	}
}

func TestConvertInvalid(t *testing.T) {
	invalids := []interface{}{
		// Incompatible struct
		struct {
			Element1 int    `yaml:"element2"`
			Element2 string `yaml:"element1"`
		}{1, "element1"},
		// Not marshable struct
		// This one panics :-|
		// struct {
		// 	Element1 func(int) int
		// }{
		// 	func(i int) int { return 0 },
		// },
	}
	for _, invalid := range invalids {
		var target yamlto
		if err := Convert(invalid, &target); err == nil {
			t.Fatalf("Expected an error converting %v to %v, got nothing", invalid, target)
		}
	}
}

func TestFilterString(t *testing.T) {
	datas := []struct {
		value    map[string][]string
		expected string
	}{
		{
			map[string][]string{},
			"{}",
		},
		{
			map[string][]string{
				"key": {},
			},
			`{"key":[]}`,
		},
		{
			map[string][]string{
				"key": {"value1", "value2"},
			},
			`{"key":["value1","value2"]}`,
		},
		{
			map[string][]string{
				"key1": {"value1", "value2"},
				"key2": {"value3", "value4"},
			},
			`{"key1":["value1","value2"],"key2":["value3","value4"]}`,
		},
	}
	for _, data := range datas {
		actual := FilterString(data.value)
		if actual != data.expected {
			t.Fatalf("Expected '%v' for %v, got '%v'", data.expected, data.value, actual)
		}
	}
}

func TestLabelFilter(t *testing.T) {
	filters := []struct {
		key      string
		value    string
		expected string
	}{
		{
			"key", "value", `{"label":["key=value"]}`,
		}, {
			"key", "", `{"label":["key="]}`,
		}, {
			"", "", `{"label":["="]}`,
		},
	}
	for _, filter := range filters {
		actual := LabelFilterString(filter.key, filter.value)
		if actual != filter.expected {
			t.Fatalf("Expected '%s for key=%s and value=%s, got %s", filter.expected, filter.key, filter.value, actual)
		}
	}
}

func TestContains(t *testing.T) {
	cases := []struct {
		collection []string
		key        string
		contains   bool
	}{
		{
			[]string{}, "", false,
		},
		{
			[]string{""}, "", true,
		},
		{
			[]string{"value1", "value2"}, "value3", false,
		},
		{
			[]string{"value1", "value2"}, "value1", true,
		},
		{
			[]string{"value1", "value2"}, "value2", true,
		},
	}
	for _, element := range cases {
		actual := Contains(element.collection, element.key)
		if actual != element.contains {
			t.Fatalf("Expected contains to be %v for %v in %v, but was %v", element.contains, element.key, element.collection, actual)
		}
	}
}
