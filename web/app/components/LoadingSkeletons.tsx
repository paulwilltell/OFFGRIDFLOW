'use client';

import { Skeleton, Stack, Box, SkeletonCircle, SkeletonText } from '@chakra-ui/react';

export function CardSkeleton() {
  return (
    <Box p={6} bg="white" borderRadius="lg" boxShadow="md">
      <Stack spacing={4}>
        <Skeleton height="20px" width="60%" />
        <Skeleton height="40px" />
        <SkeletonText mt="4" noOfLines={3} spacing="4" />
      </Stack>
    </Box>
  );
}

export function TableSkeleton({ rows = 5, columns = 4 }: { rows?: number; columns?: number }) {
  return (
    <Box>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <Stack key={rowIndex} direction="row" spacing={4} mb={3}>
          {Array.from({ length: columns }).map((_, colIndex) => (
            <Skeleton key={colIndex} height="20px" flex={1} />
          ))}
        </Stack>
      ))}
    </Box>
  );
}

export function ChartSkeleton() {
  return (
    <Box p={6} bg="white" borderRadius="lg" boxShadow="md">
      <Skeleton height="30px" width="40%" mb={4} />
      <Skeleton height="300px" />
    </Box>
  );
}

export function DashboardSkeleton() {
  return (
    <Box p={6}>
      <Stack spacing={6}>
        <Box>
          <Skeleton height="40px" width="300px" mb={2} />
          <Skeleton height="20px" width="500px" />
        </Box>
        <Stack direction={{ base: 'column', md: 'row' }} spacing={4}>
          <CardSkeleton />
          <CardSkeleton />
          <CardSkeleton />
        </Stack>
        <ChartSkeleton />
        <TableSkeleton />
      </Stack>
    </Box>
  );
}

export function ProfileSkeleton() {
  return (
    <Stack direction="row" spacing={4} align="center">
      <SkeletonCircle size="50px" />
      <Stack spacing={2} flex={1}>
        <Skeleton height="20px" width="200px" />
        <Skeleton height="15px" width="150px" />
      </Stack>
    </Stack>
  );
}

export function ListItemSkeleton({ count = 5 }: { count?: number }) {
  return (
    <Stack spacing={3}>
      {Array.from({ length: count }).map((_, index) => (
        <Stack key={index} direction="row" spacing={3} align="center">
          <SkeletonCircle size="40px" />
          <Stack spacing={2} flex={1}>
            <Skeleton height="15px" width="70%" />
            <Skeleton height="12px" width="50%" />
          </Stack>
        </Stack>
      ))}
    </Stack>
  );
}
