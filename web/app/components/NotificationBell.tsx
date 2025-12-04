'use client';

import {
  Box,
  IconButton,
  Popover,
  PopoverTrigger,
  PopoverContent,
  PopoverHeader,
  PopoverBody,
  VStack,
  HStack,
  Text,
  Badge,
  Button,
  Divider,
  useColorMode,
} from '@chakra-ui/react';
import { useState, useEffect } from 'react';

interface Notification {
  id: string;
  title: string;
  message: string;
  timestamp: Date;
  read: boolean;
  type: 'info' | 'success' | 'warning' | 'error';
}

interface NotificationBellProps {
  notifications: Notification[];
  onMarkAsRead: (id: string) => void;
  onMarkAllAsRead: () => void;
  onClear: (id: string) => void;
}

export function NotificationBell({
  notifications,
  onMarkAsRead,
  onMarkAllAsRead,
  onClear,
}: NotificationBellProps) {
  const [isOpen, setIsOpen] = useState(false);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const unreadCount = notifications.filter((n) => !n.read).length;

  const getTypeColor = (type: Notification['type']) => {
    switch (type) {
      case 'success':
        return 'green';
      case 'warning':
        return 'yellow';
      case 'error':
        return 'red';
      default:
        return 'blue';
    }
  };

  const getTypeIcon = (type: Notification['type']) => {
    switch (type) {
      case 'success':
        return '✓';
      case 'warning':
        return '⚠';
      case 'error':
        return '✕';
      default:
        return 'ℹ';
    }
  };

  return (
    <Popover isOpen={isOpen} onClose={() => setIsOpen(false)} placement="bottom-end">
      <PopoverTrigger>
        <Box position="relative" display="inline-block">
          <IconButton
            aria-label="Notifications"
            icon={<span style={{ fontSize: '20px' }}>🔔</span>}
            variant="ghost"
            onClick={() => setIsOpen(!isOpen)}
          />
          {unreadCount > 0 && (
            <Badge
              position="absolute"
              top="-2px"
              right="-2px"
              colorScheme="red"
              borderRadius="full"
              fontSize="xs"
              px={2}
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </Box>
      </PopoverTrigger>
      <PopoverContent w="400px">
        <PopoverHeader>
          <HStack justify="space-between">
            <Text fontWeight="bold">Notifications</Text>
            {unreadCount > 0 && (
              <Button size="xs" variant="ghost" onClick={onMarkAllAsRead}>
                Mark all as read
              </Button>
            )}
          </HStack>
        </PopoverHeader>
        <PopoverBody p={0}>
          {notifications.length === 0 ? (
            <Box p={6} textAlign="center">
              <Text color={isDark ? 'gray.400' : 'gray.600'}>No notifications</Text>
            </Box>
          ) : (
            <VStack spacing={0} align="stretch" maxH="400px" overflowY="auto">
              {notifications.map((notification, index) => (
                <Box key={notification.id}>
                  <HStack
                    p={3}
                    spacing={3}
                    bg={notification.read ? 'transparent' : isDark ? 'gray.700' : 'blue.50'}
                    _hover={{ bg: isDark ? 'gray.700' : 'gray.100' }}
                    cursor="pointer"
                    onClick={() => !notification.read && onMarkAsRead(notification.id)}
                  >
                    <Box
                      w="8px"
                      h="8px"
                      borderRadius="full"
                      bg={notification.read ? 'transparent' : 'blue.500'}
                      flexShrink={0}
                    />
                    <Box
                      w="32px"
                      h="32px"
                      borderRadius="md"
                      bg={`${getTypeColor(notification.type)}.500`}
                      color="white"
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      fontSize="sm"
                      fontWeight="bold"
                      flexShrink={0}
                    >
                      {getTypeIcon(notification.type)}
                    </Box>
                    <VStack flex={1} align="start" spacing={1}>
                      <Text fontSize="sm" fontWeight="medium" noOfLines={1}>
                        {notification.title}
                      </Text>
                      <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'} noOfLines={2}>
                        {notification.message}
                      </Text>
                      <Text fontSize="xs" color={isDark ? 'gray.500' : 'gray.500'}>
                        {formatTimestamp(notification.timestamp)}
                      </Text>
                    </VStack>
                    <IconButton
                      aria-label="Clear notification"
                      icon={<span>✕</span>}
                      size="xs"
                      variant="ghost"
                      onClick={(e) => {
                        e.stopPropagation();
                        onClear(notification.id);
                      }}
                      flexShrink={0}
                    />
                  </HStack>
                  {index < notifications.length - 1 && <Divider />}
                </Box>
              ))}
            </VStack>
          )}
        </PopoverBody>
      </PopoverContent>
    </Popover>
  );
}

function formatTimestamp(date: Date): string {
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return 'Just now';
  if (minutes < 60) return `${minutes}m ago`;
  if (hours < 24) return `${hours}h ago`;
  if (days < 7) return `${days}d ago`;
  return date.toLocaleDateString();
}
