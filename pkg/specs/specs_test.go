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
	assert.Equal(t, "uvw$1", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}

func TestParseValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz$1", specs[0].NewValue)
}

func TestParseOldKeyNewValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc*=def:uvw=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc(.*)$", specs[0].OldKey.String())
	assert.Equal(t, "^def$", specs[0].OldValue.String())
	assert.Equal(t, "uvw", specs[0].NewKey)
	assert.Equal(t, "xyz$1", specs[0].NewValue)
}

func TestParseOldValueNewKeyWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw*=xyz"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw$1", specs[0].NewKey)
	assert.Equal(t, "xyz", specs[0].NewValue)
}

func TestParseOldValueNewKeyValueWildcard(t *testing.T) {
	specs, err := Parse([]string{"abc=def*:uvw*=xyz*"})
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "^abc$", specs[0].OldKey.String())
	assert.Equal(t, "^def(.*)$", specs[0].OldValue.String())
	assert.Equal(t, "uvw$1", specs[0].NewKey)
	assert.Equal(t, "xyz$1", specs[0].NewValue)
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
