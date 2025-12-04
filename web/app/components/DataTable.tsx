'use client';

import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  flexRender,
  ColumnDef,
  SortingState,
  ColumnFiltersState,
  VisibilityState,
} from '@tanstack/react-table';
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Box,
  Input,
  Button,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Checkbox,
  HStack,
  IconButton,
  Select,
  Text,
  useColorMode,
} from '@chakra-ui/react';
import { useState } from 'react';

interface DataTableProps<T> {
  data: T[];
  columns: ColumnDef<T>[];
  enableSorting?: boolean;
  enableFiltering?: boolean;
  enablePagination?: boolean;
  enableColumnToggle?: boolean;
}

export function DataTable<T>({
  data,
  columns,
  enableSorting = true,
  enableFiltering = true,
  enablePagination = true,
  enableColumnToggle = true,
}: DataTableProps<T>) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [globalFilter, setGlobalFilter] = useState('');
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: enableSorting ? getSortedRowModel() : undefined,
    getFilteredRowModel: enableFiltering ? getFilteredRowModel() : undefined,
    getPaginationRowModel: enablePagination ? getPaginationRowModel() : undefined,
  });

  return (
    <Box>
      {/* Toolbar */}
      <HStack mb={4} spacing={4} flexWrap="wrap">
        {enableFiltering && (
          <Input
            placeholder="Search all columns..."
            value={globalFilter ?? ''}
            onChange={(e) => setGlobalFilter(e.target.value)}
            maxW="300px"
          />
        )}
        {enableColumnToggle && (
          <Menu closeOnSelect={false}>
            <MenuButton as={Button} size="sm">
              Columns
            </MenuButton>
            <MenuList>
              {table.getAllLeafColumns().map((column) => (
                <MenuItem key={column.id}>
                  <Checkbox
                    isChecked={column.getIsVisible()}
                    onChange={column.getToggleVisibilityHandler()}
                  >
                    {typeof column.columnDef.header === 'string'
                      ? column.columnDef.header
                      : column.id}
                  </Checkbox>
                </MenuItem>
              ))}
            </MenuList>
          </Menu>
        )}
      </HStack>

      {/* Table */}
      <Box overflowX="auto" borderRadius="lg" border="1px" borderColor={isDark ? 'gray.700' : 'gray.200'}>
        <Table variant="simple">
          <Thead bg={isDark ? 'gray.800' : 'gray.50'}>
            {table.getHeaderGroups().map((headerGroup) => (
              <Tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <Th
                    key={header.id}
                    onClick={header.column.getToggleSortingHandler()}
                    cursor={header.column.getCanSort() ? 'pointer' : 'default'}
                    userSelect="none"
                  >
                    <HStack spacing={2}>
                      <span>
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                      </span>
                      {header.column.getIsSorted() && (
                        <span>{header.column.getIsSorted() === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </HStack>
                  </Th>
                ))}
              </Tr>
            ))}
          </Thead>
          <Tbody>
            {table.getRowModel().rows.map((row) => (
              <Tr key={row.id} _hover={{ bg: isDark ? 'gray.800' : 'gray.50' }}>
                {row.getVisibleCells().map((cell) => (
                  <Td key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </Td>
                ))}
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>

      {/* Pagination */}
      {enablePagination && (
        <HStack mt={4} spacing={4} justify="space-between">
          <HStack spacing={2}>
            <Button
              size="sm"
              onClick={() => table.setPageIndex(0)}
              isDisabled={!table.getCanPreviousPage()}
            >
              {'<<'}
            </Button>
            <Button
              size="sm"
              onClick={() => table.previousPage()}
              isDisabled={!table.getCanPreviousPage()}
            >
              {'<'}
            </Button>
            <Button
              size="sm"
              onClick={() => table.nextPage()}
              isDisabled={!table.getCanNextPage()}
            >
              {'>'}
            </Button>
            <Button
              size="sm"
              onClick={() => table.setPageIndex(table.getPageCount() - 1)}
              isDisabled={!table.getCanNextPage()}
            >
              {'>>'}
            </Button>
          </HStack>
          <HStack spacing={2}>
            <Text fontSize="sm">
              Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
            </Text>
            <Select
              size="sm"
              value={table.getState().pagination.pageSize}
              onChange={(e) => table.setPageSize(Number(e.target.value))}
              w="auto"
            >
              {[10, 20, 30, 40, 50].map((pageSize) => (
                <option key={pageSize} value={pageSize}>
                  Show {pageSize}
                </option>
              ))}
            </Select>
          </HStack>
        </HStack>
      )}
    </Box>
  );
}
