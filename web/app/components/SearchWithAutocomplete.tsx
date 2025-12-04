'use client';

import {
  Input,
  List,
  ListItem,
  Box,
  VStack,
  Text,
  useColorMode,
  InputGroup,
  InputLeftElement,
  Spinner,
} from '@chakra-ui/react';
import { useState, useEffect, useRef } from 'react';

interface SearchResult {
  id: string;
  title: string;
  subtitle?: string;
  category?: string;
}

interface SearchWithAutocompleteProps {
  onSearch: (query: string) => Promise<SearchResult[]>;
  onSelect: (result: SearchResult) => void;
  placeholder?: string;
  debounceMs?: number;
}

export function SearchWithAutocomplete({
  onSearch,
  onSelect,
  placeholder = 'Search...',
  debounceMs = 300,
}: SearchWithAutocompleteProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showResults, setShowResults] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceTimer = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (debounceTimer.current) {
      clearTimeout(debounceTimer.current);
    }

    if (query.length < 2) {
      setResults([]);
      setShowResults(false);
      return;
    }

    setIsLoading(true);
    debounceTimer.current = setTimeout(async () => {
      try {
        const searchResults = await onSearch(query);
        setResults(searchResults);
        setShowResults(true);
        setSelectedIndex(-1);
      } catch (error) {
        console.error('Search error:', error);
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    }, debounceMs);

    return () => {
      if (debounceTimer.current) {
        clearTimeout(debounceTimer.current);
      }
    };
  }, [query, onSearch, debounceMs]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!showResults || results.length === 0) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex((prev) => (prev < results.length - 1 ? prev + 1 : prev));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1));
        break;
      case 'Enter':
        e.preventDefault();
        if (selectedIndex >= 0 && selectedIndex < results.length) {
          handleSelect(results[selectedIndex]);
        }
        break;
      case 'Escape':
        e.preventDefault();
        setShowResults(false);
        setSelectedIndex(-1);
        break;
    }
  };

  const handleSelect = (result: SearchResult) => {
    onSelect(result);
    setQuery(result.title);
    setShowResults(false);
    setSelectedIndex(-1);
  };

  return (
    <Box position="relative" w="100%">
      <InputGroup>
        <InputLeftElement pointerEvents="none">
          {isLoading ? <Spinner size="sm" /> : <span>🔍</span>}
        </InputLeftElement>
        <Input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => query.length >= 2 && results.length > 0 && setShowResults(true)}
          onBlur={() => setTimeout(() => setShowResults(false), 200)}
          placeholder={placeholder}
          pl="40px"
        />
      </InputGroup>

      {showResults && results.length > 0 && (
        <Box
          position="absolute"
          top="100%"
          left={0}
          right={0}
          mt={2}
          bg={isDark ? 'gray.800' : 'white'}
          borderRadius="md"
          boxShadow="lg"
          border="1px"
          borderColor={isDark ? 'gray.700' : 'gray.200'}
          zIndex={1000}
          maxH="400px"
          overflowY="auto"
        >
          <List>
            {results.map((result, index) => (
              <ListItem
                key={result.id}
                p={3}
                cursor="pointer"
                bg={selectedIndex === index ? (isDark ? 'gray.700' : 'gray.100') : 'transparent'}
                _hover={{ bg: isDark ? 'gray.700' : 'gray.100' }}
                onClick={() => handleSelect(result)}
                borderBottom={index < results.length - 1 ? '1px' : 'none'}
                borderColor={isDark ? 'gray.700' : 'gray.200'}
              >
                <VStack align="start" spacing={1}>
                  <Text fontWeight="medium" fontSize="sm">
                    {highlightMatch(result.title, query)}
                  </Text>
                  {result.subtitle && (
                    <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
                      {result.subtitle}
                    </Text>
                  )}
                  {result.category && (
                    <Text fontSize="xs" color="brand.500" fontWeight="medium">
                      {result.category}
                    </Text>
                  )}
                </VStack>
              </ListItem>
            ))}
          </List>
        </Box>
      )}
    </Box>
  );
}

function highlightMatch(text: string, query: string): React.ReactNode {
  if (!query) return text;

  const parts = text.split(new RegExp(`(${query})`, 'gi'));
  return (
    <>
      {parts.map((part, index) =>
        part.toLowerCase() === query.toLowerCase() ? (
          <Text as="span" key={index} bg="yellow.200" color="black" px={0.5}>
            {part}
          </Text>
        ) : (
          <Text as="span" key={index}>
            {part}
          </Text>
        )
      )}
    </>
  );
}
