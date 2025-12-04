'use client';

import {
  Box,
  Drawer,
  DrawerBody,
  DrawerHeader,
  DrawerOverlay,
  DrawerContent,
  DrawerCloseButton,
  VStack,
  HStack,
  IconButton,
  useDisclosure,
  useBreakpointValue,
  useColorMode,
  Button,
  Divider,
  Text,
} from '@chakra-ui/react';
import Link from 'next/link';
import { ReactNode } from 'react';

interface NavItem {
  label: string;
  href: string;
  icon: string;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', href: '/', icon: '📊' },
  { label: 'Emissions', href: '/emissions', icon: '🌍' },
  { label: 'Compliance', href: '/compliance/csrd', icon: '📋' },
  { label: 'Workflow', href: '/workflow', icon: '⚙️' },
  { label: 'Settings', href: '/settings', icon: '🔧' },
];

export function ResponsiveLayout({ children }: { children: ReactNode }) {
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { colorMode, toggleColorMode } = useColorMode();
  const isDark = colorMode === 'dark';
  const isMobile = useBreakpointValue({ base: true, md: false });

  const SidebarContent = () => (
    <VStack spacing={2} align="stretch" p={4}>
      <HStack justify="space-between" mb={4}>
        <Text fontSize="xl" fontWeight="bold">
          OffGridFlow
        </Text>
        <IconButton
          aria-label="Toggle dark mode"
          icon={<span>{isDark ? '☀️' : '🌙'}</span>}
          size="sm"
          variant="ghost"
          onClick={toggleColorMode}
        />
      </HStack>
      <Divider />
      {navItems.map((item) => (
        <Link key={item.href} href={item.href} passHref>
          <Button
            as="a"
            variant="ghost"
            justifyContent="flex-start"
            leftIcon={<span>{item.icon}</span>}
            w="full"
            onClick={isMobile ? onClose : undefined}
          >
            {item.label}
          </Button>
        </Link>
      ))}
    </VStack>
  );

  return (
    <Box minH="100vh" bg={isDark ? 'gray.900' : 'gray.50'}>
      {/* Mobile Header */}
      {isMobile && (
        <HStack
          position="sticky"
          top={0}
          zIndex={10}
          p={4}
          bg={isDark ? 'gray.800' : 'white'}
          borderBottom="1px"
          borderColor={isDark ? 'gray.700' : 'gray.200'}
          justify="space-between"
        >
          <IconButton
            aria-label="Open menu"
            icon={<span style={{ fontSize: '20px' }}>☰</span>}
            variant="ghost"
            onClick={onOpen}
          />
          <Text fontSize="lg" fontWeight="bold">
            OffGridFlow
          </Text>
          <IconButton
            aria-label="Toggle dark mode"
            icon={<span>{isDark ? '☀️' : '🌙'}</span>}
            variant="ghost"
            onClick={toggleColorMode}
          />
        </HStack>
      )}

      {/* Mobile Drawer */}
      {isMobile && (
        <Drawer isOpen={isOpen} placement="left" onClose={onClose}>
          <DrawerOverlay />
          <DrawerContent>
            <DrawerCloseButton />
            <DrawerHeader>Menu</DrawerHeader>
            <DrawerBody p={0}>
              <SidebarContent />
            </DrawerBody>
          </DrawerContent>
        </Drawer>
      )}

      {/* Desktop Layout */}
      <HStack align="start" spacing={0}>
        {!isMobile && (
          <Box
            w="250px"
            minH="100vh"
            bg={isDark ? 'gray.800' : 'white'}
            borderRight="1px"
            borderColor={isDark ? 'gray.700' : 'gray.200'}
            position="sticky"
            top={0}
          >
            <SidebarContent />
          </Box>
        )}
        <Box flex={1} p={{ base: 4, md: 6 }}>
          {children}
        </Box>
      </HStack>
    </Box>
  );
}
