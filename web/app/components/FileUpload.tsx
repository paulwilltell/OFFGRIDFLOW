'use client';

import { useDropzone } from 'react-dropzone';
import {
  Box,
  VStack,
  Text,
  Progress,
  HStack,
  IconButton,
  Image,
  useColorMode,
  List,
  ListItem,
} from '@chakra-ui/react';
import { useState, useCallback } from 'react';

interface UploadedFile {
  file: File;
  preview?: string;
  progress: number;
  status: 'uploading' | 'completed' | 'error';
}

interface FileUploadProps {
  onUpload: (files: File[]) => Promise<void>;
  accept?: Record<string, string[]>;
  maxSize?: number;
  maxFiles?: number;
  showPreview?: boolean;
}

export function FileUpload({
  onUpload,
  accept = {
    'text/csv': ['.csv'],
    'application/vnd.ms-excel': ['.xls'],
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx'],
    'application/pdf': ['.pdf'],
  },
  maxSize = 10485760, // 10MB
  maxFiles = 10,
  showPreview = true,
}: FileUploadProps) {
  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([]);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const onDrop = useCallback(
    async (acceptedFiles: File[]) => {
      const newFiles: UploadedFile[] = acceptedFiles.map((file) => ({
        file,
        preview: file.type.startsWith('image/') ? URL.createObjectURL(file) : undefined,
        progress: 0,
        status: 'uploading' as const,
      }));

      setUploadedFiles((prev) => [...prev, ...newFiles]);

      try {
        // Simulate upload progress
        for (let i = 0; i <= 100; i += 10) {
          await new Promise((resolve) => setTimeout(resolve, 100));
          setUploadedFiles((prev) =>
            prev.map((f, idx) =>
              idx >= prev.length - newFiles.length
                ? { ...f, progress: i }
                : f
            )
          );
        }

        await onUpload(acceptedFiles);

        setUploadedFiles((prev) =>
          prev.map((f, idx) =>
            idx >= prev.length - newFiles.length
              ? { ...f, status: 'completed' as const, progress: 100 }
              : f
          )
        );
      } catch (error) {
        setUploadedFiles((prev) =>
          prev.map((f, idx) =>
            idx >= prev.length - newFiles.length
              ? { ...f, status: 'error' as const }
              : f
          )
        );
      }
    },
    [onUpload]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept,
    maxSize,
    maxFiles,
  });

  const removeFile = (index: number) => {
    setUploadedFiles((prev) => {
      const file = prev[index];
      if (file.preview) {
        URL.revokeObjectURL(file.preview);
      }
      return prev.filter((_, i) => i !== index);
    });
  };

  return (
    <VStack spacing={4} align="stretch">
      <Box
        {...getRootProps()}
        p={8}
        border="2px dashed"
        borderColor={isDragActive ? 'brand.500' : isDark ? 'gray.600' : 'gray.300'}
        borderRadius="lg"
        bg={isDragActive ? (isDark ? 'brand.900' : 'brand.50') : isDark ? 'gray.800' : 'gray.50'}
        cursor="pointer"
        transition="all 0.2s"
        _hover={{
          borderColor: 'brand.500',
          bg: isDark ? 'brand.900' : 'brand.50',
        }}
      >
        <input {...getInputProps()} />
        <VStack spacing={2}>
          <Text fontSize="4xl">📁</Text>
          <Text fontWeight="medium">
            {isDragActive ? 'Drop files here' : 'Drag & drop files here, or click to select'}
          </Text>
          <Text fontSize="sm" color={isDark ? 'gray.400' : 'gray.500'}>
            Supports: CSV, Excel, PDF (max {(maxSize / 1048576).toFixed(0)}MB per file)
          </Text>
        </VStack>
      </Box>

      {uploadedFiles.length > 0 && (
        <List spacing={3}>
          {uploadedFiles.map((uploadedFile, index) => (
            <ListItem
              key={index}
              p={3}
              bg={isDark ? 'gray.800' : 'white'}
              borderRadius="md"
              boxShadow="sm"
            >
              <HStack spacing={3} align="start">
                {showPreview && uploadedFile.preview && (
                  <Image
                    src={uploadedFile.preview}
                    alt={uploadedFile.file.name}
                    boxSize="50px"
                    objectFit="cover"
                    borderRadius="md"
                  />
                )}
                <VStack flex={1} align="stretch" spacing={2}>
                  <HStack justify="space-between">
                    <Text fontSize="sm" fontWeight="medium" noOfLines={1}>
                      {uploadedFile.file.name}
                    </Text>
                    <HStack spacing={2}>
                      <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.500'}>
                        {(uploadedFile.file.size / 1024).toFixed(0)} KB
                      </Text>
                      {uploadedFile.status !== 'uploading' && (
                        <IconButton
                          aria-label="Remove file"
                          icon={<span>✕</span>}
                          size="xs"
                          variant="ghost"
                          onClick={() => removeFile(index)}
                        />
                      )}
                    </HStack>
                  </HStack>
                  {uploadedFile.status === 'uploading' && (
                    <Progress
                      value={uploadedFile.progress}
                      size="sm"
                      colorScheme="brand"
                      borderRadius="full"
                    />
                  )}
                  {uploadedFile.status === 'completed' && (
                    <Text fontSize="xs" color="green.500">
                      ✓ Upload complete
                    </Text>
                  )}
                  {uploadedFile.status === 'error' && (
                    <Text fontSize="xs" color="red.500">
                      ✗ Upload failed
                    </Text>
                  )}
                </VStack>
              </HStack>
            </ListItem>
          ))}
        </List>
      )}
    </VStack>
  );
}
