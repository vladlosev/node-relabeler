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
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}

func TestParseKeyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw*=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].OldKey.String())
	assert.Equal(t, "^def$", specs[0].OldValue.String())
	assert.Equal(t, "uvw*", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}

func TestParseValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz*", specs[0].NewValue)
}

func TestParseOldKeyNewValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].OldKey.String())
	assert.Equal(t, "^def$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz*", specs[0].NewValue)
}

func TestParseOldValueNewKeyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw*=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw*", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}

func TestParseOldValueNewKeyValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw*=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw*", specs[0].NewKey)
	assert.Equal(t, "xyz*", specs[0].NewValue)
}
func TestParseOldKeyOnlyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].OldKey.String())
	assert.Equal(t, "^def$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}
func TestParseOldValueOnlyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}

func TestParseEmptyFails(t *testing.T) {
	_, err := Parse([]string{})
	require.Error(t, err)
	assert.Regexp(t, "At least one", err.Error())
}

func TestParseNewKeyWildcardFails(t *testing.T) {
	_, err := Parse([]string{"abc=def:uvw*=xyz"})
	require.Error(t, err)
	assert.Regexp(
		t,
		"cannot appear in new label without appearing in the old one",
		err.Error())
}

func TestParseNewValueWildcardFails(t *testing.T) {
	_, err := Parse([]string{"abc=def:uvw=xyz*"})
	require.Error(t, err)
	assert.Regexp(
		t,
		"cannot appear in new label without appearing in the old one",
		err.Error())
}

func TestParseOldKeyValueWildcardFails(t *testing.T) {
	_, err := Parse([]string{"abc*=def*:uvw=xyz"})
	require.Error(t, err)
	assert.Regexp(
		t,
		"oldkey=oldvalue pair should contain no more than a single",
		err.Error())
}

func TestApplyToSimpleEmpty(t *testing.T) {
	specs, err := Parse([]string{"abc=def:uvw=xyz"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{})
	assert.Empty(t, results)
}

func TestApplyToSimpleMismatch(t *testing.T) {
	specs, err := Parse([]string{"abc=def:uvw=xyz"})
	require.NoError(t, err)
	results := specs.ApplyTo(map[string]string{"abd": "123"})
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
