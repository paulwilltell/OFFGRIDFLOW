'use client';

import { Box, VStack, Text, Button, Icon, useColorMode } from '@chakra-ui/react';
import { ReactNode } from 'react';

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  return (
    <Box
      textAlign="center"
      py={12}
      px={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
    >
      <VStack spacing={4}>
        {icon && (
          <Box fontSize="4xl" color={isDark ? 'gray.400' : 'gray.500'}>
            {icon}
          </Box>
        )}
        <Text fontSize="xl" fontWeight="bold" color={isDark ? 'gray.200' : 'gray.700'}>
          {title}
        </Text>
        <Text fontSize="md" color={isDark ? 'gray.400' : 'gray.500'} maxW="md">
          {description}
        </Text>
        {action && (
          <Button colorScheme="brand" onClick={action.onClick} mt={4}>
            {action.label}
          </Button>
        )}
      </VStack>
    </Box>
  );
}

export function NoDataEmptyState({ onRefresh }: { onRefresh?: () => void }) {
  return (
    <EmptyState
      icon={<span>📊</span>}
      title="No Data Available"
      description="There's no data to display yet. Start by adding your first emission record or connecting a data source."
      action={
        onRefresh
          ? {
              label: 'Refresh Data',
              onClick: onRefresh,
            }
          : undefined
      }
    />
  );
}

export function NoResultsEmptyState({ onClearFilters }: { onClearFilters?: () => void }) {
  return (
    <EmptyState
      icon={<span>🔍</span>}
      title="No Results Found"
      description="We couldn't find any results matching your search criteria. Try adjusting your filters or search terms."
      action={
        onClearFilters
          ? {
              label: 'Clear Filters',
              onClick: onClearFilters,
            }
          : undefined
      }
    />
  );
}

export function ErrorEmptyState({ onRetry }: { onRetry?: () => void }) {
  return (
    <EmptyState
      icon={<span>⚠️</span>}
      title="Something Went Wrong"
      description="We encountered an error while loading your data. Please try again or contact support if the problem persists."
      action={
        onRetry
          ? {
              label: 'Try Again',
              onClick: onRetry,
            }
          : undefined
      }
    />
  );
}
