# CLI Wordlist Management Guide

## Overview

The Go Distributed Hashcat server now includes a comprehensive CLI for managing wordlist files, specifically optimized for large files with 1 million+ words. This CLI provides fast, efficient processing with progress reporting and configurable options.

## Features

- **Optimized Processing**: Buffered I/O and efficient file handling for large files
- **Progress Reporting**: Real-time progress updates during upload
- **Configurable Chunking**: Adjustable chunk size for memory-efficient processing
- **Word Counting**: Automatic word count calculation (can be disabled for speed)
- **File Validation**: Automatic file format detection and validation
- **Memory Efficient**: Processes files in chunks to handle very large wordlists

## Commands

### `wordlist upload [path]`

Upload a wordlist file with optimized processing.

#### Parameters

- `path` (required): Path to the wordlist file to upload
- `--name` (optional): Custom name for the wordlist
- `--count` (optional): Enable word counting (default: true)
- `--chunk` (optional): Chunk size for processing in MB (default: 10)

#### Examples

```bash
# Basic upload
./server wordlist upload /path/to/rockyou.txt

# Upload with custom name
./server wordlist upload /path/to/wordlist.txt --name "custom_name"

# Upload large files with custom chunk size
./server wordlist upload /path/to/large.txt --chunk 50

# Fast upload without word counting
./server wordlist upload /path/to/wordlist.txt --count=false

# Upload with custom name and word counting enabled
./server wordlist upload /home/user/Downloads/rockyou.txt --name "rockyou_2021" --count=true

# Upload with custom name and word counting disabled (faster)
./server wordlist upload /home/user/Downloads/rockyou.txt --name "rockyou_2021" --count=false
```

### `wordlist list`

List all uploaded wordlists with their details.

#### Examples

```bash
./server wordlist list
```

#### Output Format

```
Uploaded Wordlists:
---------------------
ID: 2bdac58a-2d2b-47f8-bcb5-2441281bae3f
Name: 2bdac58a-2d2b-47f8-bcb5-2441281bae3f.txt
Size: 64 B
Word Count: 10
Created At: 2025-08-19T20:35:16-07:00
---------------------
```

### `wordlist delete [id]`

Delete a wordlist by its UUID.

#### Examples

```bash
./server wordlist delete 2bdac58a-2d2b-47f8-bcb5-2441281bae3f
```

## Performance Optimization

### For Large Wordlists (1M+ words)

- **Chunk Size**: Use `--chunk 50` for files larger than 1GB
- **Word Counting**: Disable with `--count=false` for fastest uploads
- **Memory Management**: Larger chunk sizes reduce memory overhead
- **Progress Monitoring**: Smaller chunks provide more frequent progress updates

### Recommended Settings

| File Size | Chunk Size | Word Counting | Use Case |
|-----------|------------|---------------|----------|
| < 100MB  | 10MB       | Enabled       | Development/Testing |
| 100MB-1GB| 25MB       | Enabled       | Medium wordlists |
| 1GB-5GB  | 50MB       | Optional      | Large wordlists |
| > 5GB    | 100MB      | Disabled      | Very large wordlists |

## Technical Details

### Progress Reporting

The CLI provides real-time progress updates during upload:

```
Starting wordlist upload: large_test
File size: 965.8 KB
Chunk size: 50.0 MB
Word counting: true
✅ Wordlist uploaded successfully!
📁 ID: 4d05cadc-6596-4089-a117-449eadb44f3c
📝 Name: 4d05cadc-6596-4089-a117-449eadb44f3c.txt
📊 Size: 965.8 KB
🔢 Word count: 100,000
💾 Path: uploads/wordlists/4d05cadc-6596-4089-a117-449eadb44f3c.txt
```

### File Processing

- **Chunked Reading**: Files are processed in configurable chunks
- **Buffered I/O**: Uses Go's buffered I/O for efficient file handling
- **Word Counting**: Line-by-line processing with empty line filtering
- **Error Handling**: Comprehensive error handling with cleanup on failure

### Storage

- **Directory Structure**: `uploads/wordlists/` for file storage
- **Naming Convention**: UUID-based filenames with original extensions
- **Database Records**: Full metadata storage including word counts
- **Cleanup**: Automatic cleanup of failed uploads

### Database Storage Examples

#### Example 1: With Word Counting Enabled
```bash
./server wordlist upload /home/user/Downloads/rockyou.txt --name "rockyou_2021" --count=true
```

**Database Result**:
```sql
id: 4d05cadc-6596-4089-a117-449eadb44f3c
name: 4d05cadc-6596-4089-a117-449eadb44f3c.txt
orig_name: rockyou.txt
path: uploads/wordlists/4d05cadc-6596-4089-a117-449eadb44f3c.txt
size: 1434439168
word_count: 14344391  -- Word count calculated
created_at: 2025-08-19 20:37:16
```

#### Example 2: With Word Counting Disabled
```bash
./server wordlist upload /home/user/Downloads/rockyou.txt --name "rockyou_2021" --count=false
```

**Database Result**:
```sql
id: 4d05cadc-6596-4089-a117-449eadb44f3c
name: 4d05cadc-6596-4089-a117-449eadb44f3c.txt
orig_name: rockyou.txt
path: uploads/wordlists/4d05cadc-6596-4089-a117-449eadb44f3c.txt
size: 1434439168
word_count: NULL  -- Word count not calculated (faster upload)
created_at: 2025-08-19 20:37:16
```

**Key Differences**:
- **`--count=true`**: Slower upload but provides word count in database
- **`--count=false`**: Faster upload but word count will be NULL
- **File storage**: Both commands store the same file information
- **Performance**: `--count=false` is significantly faster for large files

## Best Practices

### Upload Performance

1. **Chunk Size**: Start with default (10MB) and adjust based on file size
2. **Word Counting**: Disable for files > 10M words if speed is priority
3. **File Format**: Ensure clean text format (one word per line)
4. **Disk Space**: Verify sufficient space in uploads directory

### File Management

1. **Naming**: Use descriptive custom names for easy identification
2. **Organization**: Group related wordlists with consistent naming
3. **Cleanup**: Regularly remove unused wordlists to save space
4. **Backup**: Important wordlists should be backed up externally

### Monitoring

1. **Progress**: Monitor upload progress for large files
2. **Resources**: Watch memory and disk usage during uploads
3. **Errors**: Check logs for any upload failures
4. **Performance**: Track upload times for optimization

## Troubleshooting

### Common Issues

- **File Not Found**: Verify file path and permissions
- **Insufficient Space**: Check available disk space in uploads directory
- **Permission Denied**: Ensure write access to uploads directory
- **Invalid UUID**: Use correct UUID format for delete operations

### Performance Issues

- **Slow Uploads**: Increase chunk size or disable word counting
- **Memory Usage**: Reduce chunk size for memory-constrained systems
- **Progress Updates**: Smaller chunks provide more frequent updates

## Examples

### Complete Workflow

```bash
# 1. Upload a large wordlist
./server wordlist upload /path/to/rockyou.txt --name "rockyou_2021" --chunk 50

# 2. List all wordlists
./server wordlist list

# 3. Delete a wordlist
./server wordlist delete <uuid_from_list>
```

### Batch Processing

```bash
# Upload multiple wordlists
for file in wordlists/*.txt; do
    name=$(basename "$file" .txt)
    ./server wordlist upload "$file" --name "$name" --chunk 25
done
```

### Performance Testing

```bash
# Test with different chunk sizes
./server wordlist upload large.txt --chunk 10 --count=true
./server wordlist upload large.txt --chunk 50 --count=true
./server wordlist upload large.txt --chunk 100 --count=false
```

### Word Counting Comparison

```bash
# Upload with word counting enabled (slower but provides word count)
./server wordlist upload /home/user/Downloads/rockyou.txt --name "rockyou_2021" --count=true

# Upload with word counting disabled (faster but no word count)
./server wordlist upload /home/user/Downloads/rockyou.txt --name "rockyou_2021" --count=false

# Compare upload times and database results
./server wordlist list
```

## Integration

The CLI integrates seamlessly with the existing server architecture:

- **Database**: Uses existing wordlist repository and models
- **Storage**: Follows established file storage patterns
- **API**: CLI operations are independent of HTTP API
- **Monitoring**: Integrates with existing logging and error handling

## Future Enhancements

Planned improvements for the CLI:

- **Parallel Processing**: Multi-threaded uploads for very large files
- **Compression**: Automatic compression/decompression support
- **Validation**: Enhanced file format validation
- **Resume**: Resume interrupted uploads
- **Batch Operations**: Upload multiple files simultaneously
