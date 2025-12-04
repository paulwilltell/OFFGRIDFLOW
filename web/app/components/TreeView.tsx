'use client';

import {
  Box,
  HStack,
  VStack,
  Text,
  Collapse,
  IconButton,
  Checkbox,
  useColorMode,
} from '@chakra-ui/react';
import { useState } from 'react';

export interface TreeNode {
  id: string;
  label: string;
  children?: TreeNode[];
  data?: any;
}

interface TreeViewProps {
  data: TreeNode[];
  onSelect?: (node: TreeNode) => void;
  selectable?: boolean;
  multiSelect?: boolean;
  defaultExpanded?: string[];
}

export function TreeView({
  data,
  onSelect,
  selectable = false,
  multiSelect = false,
  defaultExpanded = [],
}: TreeViewProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set(defaultExpanded));
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const toggleExpanded = (nodeId: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  const toggleSelected = (node: TreeNode) => {
    if (!selectable) return;

    setSelected((prev) => {
      const next = new Set(prev);
      if (multiSelect) {
        if (next.has(node.id)) {
          next.delete(node.id);
        } else {
          next.add(node.id);
        }
      } else {
        next.clear();
        next.add(node.id);
      }
      return next;
    });

    if (onSelect) {
      onSelect(node);
    }
  };

  const renderNode = (node: TreeNode, level: number = 0) => {
    const hasChildren = node.children && node.children.length > 0;
    const isExpanded = expanded.has(node.id);
    const isSelected = selected.has(node.id);

    return (
      <Box key={node.id}>
        <HStack
          spacing={2}
          pl={level * 6}
          py={2}
          px={3}
          cursor="pointer"
          borderRadius="md"
          bg={isSelected ? (isDark ? 'brand.700' : 'brand.100') : 'transparent'}
          _hover={{ bg: isDark ? 'gray.700' : 'gray.100' }}
          onClick={() => selectable && toggleSelected(node)}
        >
          {hasChildren && (
            <IconButton
              aria-label={isExpanded ? 'Collapse' : 'Expand'}
              icon={<span>{isExpanded ? '▼' : '▶'}</span>}
              size="xs"
              variant="ghost"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpanded(node.id);
              }}
            />
          )}
          {!hasChildren && <Box w="24px" />}
          {selectable && multiSelect && (
            <Checkbox
              isChecked={isSelected}
              onChange={() => toggleSelected(node)}
              onClick={(e) => e.stopPropagation()}
            />
          )}
          <Text
            fontSize="sm"
            fontWeight={isSelected ? 'bold' : 'normal'}
            flex={1}
            color={isSelected ? (isDark ? 'white' : 'brand.700') : undefined}
          >
            {node.label}
          </Text>
        </HStack>
        {hasChildren && (
          <Collapse in={isExpanded} animateOpacity>
            <VStack align="stretch" spacing={0}>
              {node.children!.map((child) => renderNode(child, level + 1))}
            </VStack>
          </Collapse>
        )}
      </Box>
    );
  };

  return (
    <Box
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
      borderRadius="md"
      p={2}
      bg={isDark ? 'gray.800' : 'white'}
    >
      <VStack align="stretch" spacing={0}>
        {data.map((node) => renderNode(node))}
      </VStack>
    </Box>
  );
}
