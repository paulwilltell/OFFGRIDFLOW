'use client';

import {
  Box,
  SimpleGrid,
  VStack,
  HStack,
  Text,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Button,
  useColorMode,
  Divider,
  Badge,
  Progress,
} from '@chakra-ui/react';

interface KPICardProps {
  title: string;
  value: string | number;
  change?: number;
  trend?: 'up' | 'down';
  subtitle?: string;
  icon?: string;
  color?: string;
}

export function KPICard({ title, value, change, trend, subtitle, icon, color = 'brand' }: KPICardProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
      _hover={{ boxShadow: 'lg', transform: 'translateY(-2px)', transition: 'all 0.2s' }}
    >
      <Stat>
        <HStack justify="space-between" mb={2}>
          <StatLabel fontSize="sm" color={isDark ? 'gray.400' : 'gray.600'}>
            {title}
          </StatLabel>
          {icon && <Text fontSize="2xl">{icon}</Text>}
        </HStack>
        <StatNumber fontSize="3xl" fontWeight="bold" color={`${color}.500`}>
          {value}
        </StatNumber>
        {(change !== undefined || subtitle) && (
          <StatHelpText mt={2}>
            {change !== undefined && trend && (
              <HStack spacing={1}>
                <StatArrow type={trend === 'up' ? 'increase' : 'decrease'} />
                <Text>{Math.abs(change)}%</Text>
              </HStack>
            )}
            {subtitle && <Text fontSize="xs">{subtitle}</Text>}
          </StatHelpText>
        )}
      </Stat>
    </Box>
  );
}

export function ExecutiveSummaryWidget() {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <Text fontSize="xl" fontWeight="bold" mb={4}>
        Executive Summary
      </Text>
      <VStack spacing={4} align="stretch">
        <Box>
          <Text fontSize="sm" fontWeight="medium" mb={2}>
            Carbon Footprint
          </Text>
          <Text fontSize="2xl" fontWeight="bold" color="brand.500">
            12,450 tCO2e
          </Text>
          <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
            6.2% decrease from last quarter
          </Text>
        </Box>
        <Divider />
        <Box>
          <Text fontSize="sm" fontWeight="medium" mb={2}>
            Compliance Status
          </Text>
          <HStack spacing={2} mb={2}>
            <Badge colorScheme="green">CSRD Ready</Badge>
            <Badge colorScheme="yellow">SEC In Progress</Badge>
          </HStack>
          <Progress value={75} colorScheme="green" size="sm" borderRadius="full" />
        </Box>
        <Divider />
        <Box>
          <Text fontSize="sm" fontWeight="medium" mb={2}>
            Key Metrics
          </Text>
          <VStack spacing={2} align="stretch">
            <HStack justify="space-between">
              <Text fontSize="xs">Energy Efficiency</Text>
              <Text fontSize="xs" fontWeight="bold" color="green.500">
                +12%
              </Text>
            </HStack>
            <HStack justify="space-between">
              <Text fontSize="xs">Renewable Energy</Text>
              <Text fontSize="xs" fontWeight="bold" color="brand.500">
                45%
              </Text>
            </HStack>
            <HStack justify="space-between">
              <Text fontSize="xs">Data Quality Score</Text>
              <Text fontSize="xs" fontWeight="bold" color="blue.500">
                8.7/10
              </Text>
            </HStack>
          </VStack>
        </Box>
      </VStack>
    </Box>
  );
}

export function RecentActivityFeed() {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const activities = [
    { id: 1, action: 'Uploaded utility bill', time: '2 hours ago', icon: '📄' },
    { id: 2, action: 'Completed CSRD report', time: '5 hours ago', icon: '✅' },
    { id: 3, action: 'Added new facility', time: '1 day ago', icon: '🏭' },
    { id: 4, action: 'Data source connected', time: '2 days ago', icon: '🔌' },
    { id: 5, action: 'Team member invited', time: '3 days ago', icon: '👤' },
  ];

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <Text fontSize="xl" fontWeight="bold" mb={4}>
        Recent Activity
      </Text>
      <VStack spacing={3} align="stretch">
        {activities.map((activity) => (
          <HStack key={activity.id} spacing={3}>
            <Box
              w="32px"
              h="32px"
              borderRadius="full"
              bg={isDark ? 'gray.700' : 'gray.100'}
              display="flex"
              alignItems="center"
              justifyContent="center"
              fontSize="sm"
            >
              {activity.icon}
            </Box>
            <VStack align="start" spacing={0} flex={1}>
              <Text fontSize="sm" fontWeight="medium">
                {activity.action}
              </Text>
              <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
                {activity.time}
              </Text>
            </VStack>
          </HStack>
        ))}
      </VStack>
    </Box>
  );
}

export function ComplianceDeadlinesWidget() {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const deadlines = [
    { name: 'CSRD Report Submission', date: 'Dec 31, 2024', priority: 'high', daysLeft: 30 },
    { name: 'SEC Climate Disclosure', date: 'Jan 15, 2025', priority: 'medium', daysLeft: 45 },
    { name: 'CBAM Report Q4', date: 'Feb 1, 2025', priority: 'medium', daysLeft: 62 },
  ];

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return 'red';
      case 'medium':
        return 'yellow';
      default:
        return 'green';
    }
  };

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <Text fontSize="xl" fontWeight="bold" mb={4}>
        Upcoming Deadlines
      </Text>
      <VStack spacing={3} align="stretch">
        {deadlines.map((deadline, index) => (
          <Box
            key={index}
            p={3}
            borderRadius="md"
            bg={isDark ? 'gray.700' : 'gray.50'}
            borderLeft="4px"
            borderColor={`${getPriorityColor(deadline.priority)}.500`}
          >
            <HStack justify="space-between" mb={1}>
              <Text fontSize="sm" fontWeight="medium">
                {deadline.name}
              </Text>
              <Badge colorScheme={getPriorityColor(deadline.priority)} fontSize="xs">
                {deadline.daysLeft} days
              </Badge>
            </HStack>
            <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
              Due: {deadline.date}
            </Text>
          </Box>
        ))}
      </VStack>
      <Button size="sm" variant="ghost" mt={3} w="full">
        View All Deadlines →
      </Button>
    </Box>
  );
}

export function DataSourceHealthWidget() {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const dataSources = [
    { name: 'Utility Bills API', status: 'healthy', uptime: 99.9 },
    { name: 'SAP Connector', status: 'healthy', uptime: 98.5 },
    { name: 'Excel Import', status: 'warning', uptime: 85.2 },
    { name: 'IoT Sensors', status: 'healthy', uptime: 97.8 },
  ];

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'green';
      case 'warning':
        return 'yellow';
      case 'error':
        return 'red';
      default:
        return 'gray';
    }
  };

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <Text fontSize="xl" fontWeight="bold" mb={4}>
        Data Source Health
      </Text>
      <VStack spacing={3} align="stretch">
        {dataSources.map((source, index) => (
          <HStack key={index} justify="space-between">
            <HStack spacing={2} flex={1}>
              <Box
                w="8px"
                h="8px"
                borderRadius="full"
                bg={`${getStatusColor(source.status)}.500`}
              />
              <Text fontSize="sm">{source.name}</Text>
            </HStack>
            <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
              {source.uptime}%
            </Text>
          </HStack>
        ))}
      </VStack>
    </Box>
  );
}

export function CarbonReductionTargetsWidget() {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const targets = [
    { name: '2025 Target', current: 12450, target: 11000, progress: 88 },
    { name: '2030 Net Zero', current: 12450, target: 0, progress: 15 },
  ];

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <Text fontSize="xl" fontWeight="bold" mb={4}>
        Carbon Reduction Targets
      </Text>
      <VStack spacing={4} align="stretch">
        {targets.map((target, index) => (
          <Box key={index}>
            <HStack justify="space-between" mb={2}>
              <Text fontSize="sm" fontWeight="medium">
                {target.name}
              </Text>
              <Text fontSize="sm" color="brand.500" fontWeight="bold">
                {target.progress}%
              </Text>
            </HStack>
            <Progress
              value={target.progress}
              colorScheme={target.progress >= 75 ? 'green' : 'yellow'}
              size="sm"
              borderRadius="full"
            />
            <HStack justify="space-between" mt={1}>
              <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
                Current: {target.current.toLocaleString()} tCO2e
              </Text>
              <Text fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
                Target: {target.target.toLocaleString()} tCO2e
              </Text>
            </HStack>
          </Box>
        ))}
      </VStack>
    </Box>
  );
}

export function QuickActionsWidget() {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const actions = [
    { label: 'Upload Data', icon: '📤', color: 'blue' },
    { label: 'Generate Report', icon: '📊', color: 'green' },
    { label: 'Add Facility', icon: '🏭', color: 'purple' },
    { label: 'Invite Team', icon: '👥', color: 'orange' },
  ];

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <Text fontSize="xl" fontWeight="bold" mb={4}>
        Quick Actions
      </Text>
      <SimpleGrid columns={2} spacing={3}>
        {actions.map((action, index) => (
          <Button
            key={index}
            variant="outline"
            colorScheme={action.color}
            leftIcon={<span>{action.icon}</span>}
            h="60px"
            fontSize="sm"
            flexDirection="column"
          >
            {action.label}
          </Button>
        ))}
      </SimpleGrid>
    </Box>
  );
}
