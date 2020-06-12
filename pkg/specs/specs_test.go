package specs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSimple(t *testing.T) {
	specs, err := Parse([]string{"abc=def:uvw=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw", specs[0].newKey)
	assert.Equal(t, "xyz", specs[0].newValue)
}

func TestParseKeyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw*=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw*", specs[0].newKey)
	assert.Equal(t, "xyz", specs[0].newValue)
}

func TestParseValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def(.*)$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw", specs[0].newKey)
	assert.Equal(t, "xyz*", specs[0].newValue)
}

func TestParseOldKeyNewValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw", specs[0].newKey)
	assert.Equal(t, "xyz*", specs[0].newValue)
}

func TestParseOldValueNewKeyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw*=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def(.*)$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw*", specs[0].newKey)
	assert.Equal(t, "xyz", specs[0].newValue)
}

func TestParseOldValueNewKeyValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw*=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def(.*)$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw*", specs[0].newKey)
	assert.Equal(t, "xyz*", specs[0].newValue)
}
func TestParseOldKeyOnlyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw", specs[0].newKey)
	assert.Equal(t, "xyz", specs[0].newValue)
}
func TestParseOldValueOnlyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].oldKeyRegexp.String())
	assert.Equal(t, "^def(.*)$", specs[0].oldValueRegexp.String())
	assert.Equal(t, "uvw", specs[0].newKey)
	assert.Equal(t, "xyz", specs[0].newValue)
}

func TestParseLabelSpecFailures(t *testing.T) {
	testData := []struct {
		name    string
		specs   []string
		message string
	}{
		{
			"Empty",
			[]string{},
			"At least one",
		},
		{
			"TooManyLabelSpecs",
			[]string{"abcd=def:ghi=jkl:uvw=xyz:"},
			"Specs must be in the form",
		},
		{
			"NotEnoughLabelSpecs",
			[]string{"uvw=xyz"},
			"Specs must be in the form",
		},
		{
			"OldLabelSpecTooManyParts",
			[]string{"abc=def=hjk:uvw=xyz"},
			"Specs must be in the form",
		},
		{
			"NewLabelSpecTooManyParts",
			[]string{"abc=def:uvw=xyz=123"},
			"Specs must be in the form",
		},
		{
			"NewKeyWildcardOnly",
			[]string{"abc=def:uvw*=xyz"},
			"cannot appear in new label without appearing in the old one",
		},
		{
			"NewValueWildcardOnly",
			[]string{"abc=def:uvw=xyz*"},
			"cannot appear in new label without appearing in the old one",
		},
		{
			"BothOldKeyAndValueWildcard",
			[]string{"abc*=def*:uvw=xyz"},
			"oldkey=oldvalue pair should contain no more than a single",
		},
		{
			"AutoRecursiveKeyWildcard",
			[]string{"abc*=123:abcd*=123"},
			"newkey=newvalue pair must not match pattern in oldkey=oldvalue",
		},
		{
			"AutoRecursiveValueWildcard",
			[]string{"abc=*:abc=*x"},
			"newkey=newvalue pair must not match pattern in oldkey=oldvalue",
		},
	}
	for _, testItem := range testData {
		t.Run(testItem.name, func(t *testing.T) {
			_, err := Parse(testItem.specs)
			require.Error(t, err)
			assert.Regexp(t, testItem.message, err.Error())
		})
	}
}

func TestApplyToSimpleEmpty(t *testing.T) {
	specs, err := Parse([]string{"abc=def:uvw=xyz"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{})
	assert.Empty(t, results)
}

func TestApplyToSimpleKeyMismatch(t *testing.T) {
	specs, err := Parse([]string{"abc=def:uvw=xyz"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abd": "123"})
	assert.Empty(t, results)
}

func TestApplyToSimpleValueMismatch(t *testing.T) {
	specs, err := Parse([]string{"abc=def:uvw=xyz"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc": "123"})
	assert.Empty(t, results)
}

func TestApplyToSimpleReplaceValue(t *testing.T) {
	specs, err := Parse([]string{"abc=def:abc=xyz"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc": "def"})
	assert.Equal(t, results, map[string]string{"abc": "xyz"})
}
func TestApplyToSimpleReplaceKey(t *testing.T) {
	specs, err := Parse([]string{"abc=def:pqr=def"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc": "def"})
	assert.Equal(t, results, map[string]string{"pqr": "def"})
}

func TestApplyToSimpleReplaceKeyValue(t *testing.T) {
	specs, err := Parse([]string{"abc=def:pqr=stu"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc": "def"})
	assert.Equal(t, results, map[string]string{"pqr": "stu"})
}
func TestApplyToWildcardKey(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:pqr*=def"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc123": "def"})
	assert.Equal(t, results, map[string]string{"pqr123": "def"})
}

func TestApplyToWildcardValue(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:pqr=xyz*"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc": "def123"})
	assert.Equal(t, results, map[string]string{"pqr": "xyz123"})
}

func TestApplyToWildcardKeyReplaceKeyValue(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:pqr*=def*"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc123": "def"})
	assert.Equal(t, results, map[string]string{"pqr123": "def123"})
}

func TestApplyToWildcardValueReplaceKeyValue(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:pqr*=def*"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abc": "def123"})
	assert.Equal(t, results, map[string]string{"pqr123": "def123"})
}
