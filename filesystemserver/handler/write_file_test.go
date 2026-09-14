package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWriteFile(t *testing.T) {
	// Setup a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a handler with the temp dir as an allowed path
	allowedDirs := resolveAllowedDirs(t, tmpDir)
	fsHandler, err := NewFilesystemHandler(allowedDirs)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("write to an existing directory", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "existing_dir.txt")
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": "hello world",
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.False(t, res.IsError)

		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("write with non-existent parent directories", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "a", "b", "c", "new.txt")
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": "nested",
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.False(t, res.IsError)

		// Verify parent directories were created
		info, err := os.Stat(filepath.Dir(filePath))
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Verify the file was created
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, "nested", string(data))
	})

	t.Run("write to a non-allowed directory does not create it", func(t *testing.T) {
		otherDir := t.TempDir()
		filePath := filepath.Join(otherDir, "new_dir", "file.txt")
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    filePath,
					"content": "denied",
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.True(t, res.IsError)

		// Verify no directories were created outside allowed dirs
		_, err = os.Stat(filepath.Join(otherDir, "new_dir"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("write to a directory path is an error", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"path":    tmpDir,
					"content": "nope",
				},
			},
		}

		res, err := fsHandler.HandleWriteFile(ctx, req)
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}
