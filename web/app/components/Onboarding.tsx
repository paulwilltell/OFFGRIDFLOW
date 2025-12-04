'use client';

import { Steps } from 'intro.js-react';
import 'intro.js/introjs.css';
import { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  VStack,
  HStack,
  Text,
  Checkbox,
  Progress,
  Box,
  useColorMode,
  List,
  ListItem,
  ListIcon,
} from '@chakra-ui/react';

interface OnboardingStep {
  element: string;
  intro: string;
  title?: string;
}

const onboardingSteps: OnboardingStep[] = [
  {
    element: '.dashboard-header',
    title: 'Welcome to OffGridFlow!',
    intro: 'This is your carbon accounting and compliance platform. Let me show you around.',
  },
  {
    element: '.kpi-cards',
    title: 'Key Metrics',
    intro: 'Here you can see your most important emissions metrics at a glance.',
  },
  {
    element: '.emissions-chart',
    title: 'Emissions Visualization',
    intro: 'View your emissions trends over time with interactive charts.',
  },
  {
    element: '.compliance-status',
    title: 'Compliance Tracking',
    intro: 'Monitor your compliance status across different frameworks like CSRD, SEC, and CBAM.',
  },
  {
    element: '.quick-actions',
    title: 'Quick Actions',
    intro: 'Use these shortcuts to perform common tasks like uploading data or generating reports.',
  },
];

interface OnboardingTourProps {
  enabled: boolean;
  onComplete: () => void;
}

export function OnboardingTour({ enabled, onComplete }: OnboardingTourProps) {
  const [stepsEnabled, setStepsEnabled] = useState(enabled);
  const { colorMode } = useColorMode();

  useEffect(() => {
    setStepsEnabled(enabled);
  }, [enabled]);

  return (
    <Steps
      enabled={stepsEnabled}
      steps={onboardingSteps}
      initialStep={0}
      onExit={() => {
        setStepsEnabled(false);
        onComplete();
      }}
      options={{
        showProgress: true,
        showBullets: false,
        exitOnOverlayClick: false,
        doneLabel: 'Finish',
        nextLabel: 'Next',
        prevLabel: 'Back',
        skipLabel: 'Skip',
      }}
    />
  );
}

interface SetupChecklistItem {
  id: string;
  label: string;
  completed: boolean;
  description: string;
}

interface SetupChecklistProps {
  items: SetupChecklistItem[];
  onItemToggle: (id: string) => void;
  onComplete: () => void;
}

export function SetupChecklist({ items, onItemToggle, onComplete }: SetupChecklistProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';
  const completedCount = items.filter((item) => item.completed).length;
  const progress = (completedCount / items.length) * 100;
  const allCompleted = completedCount === items.length;

  return (
    <Box
      p={6}
      bg={isDark ? 'gray.800' : 'white'}
      borderRadius="lg"
      boxShadow="md"
      border="1px"
      borderColor={isDark ? 'gray.700' : 'gray.200'}
    >
      <VStack spacing={4} align="stretch">
        <HStack justify="space-between">
          <Text fontSize="xl" fontWeight="bold">
            Setup Checklist
          </Text>
          <Text fontSize="sm" color={isDark ? 'gray.400' : 'gray.600'}>
            {completedCount} of {items.length} completed
          </Text>
        </HStack>

        <Progress
          value={progress}
          colorScheme={allCompleted ? 'green' : 'brand'}
          borderRadius="full"
          size="sm"
        />

        <List spacing={3}>
          {items.map((item) => (
            <ListItem
              key={item.id}
              p={3}
              borderRadius="md"
              bg={isDark ? 'gray.700' : 'gray.50'}
              cursor="pointer"
              onClick={() => onItemToggle(item.id)}
              opacity={item.completed ? 0.7 : 1}
            >
              <HStack spacing={3} align="start">
                <Checkbox
                  isChecked={item.completed}
                  onChange={() => onItemToggle(item.id)}
                  size="lg"
                  colorScheme="green"
                />
                <VStack align="start" spacing={1} flex={1}>
                  <Text
                    fontWeight="medium"
                    textDecoration={item.completed ? 'line-through' : 'none'}
                  >
                    {item.label}
                  </Text>
                  <Text fontSize="sm" color={isDark ? 'gray.400' : 'gray.600'}>
                    {item.description}
                  </Text>
                </VStack>
              </HStack>
            </ListItem>
          ))}
        </List>

        {allCompleted && (
          <Button colorScheme="green" onClick={onComplete}>
            🎉 Complete Setup
          </Button>
        )}
      </VStack>
    </Box>
  );
}

interface WelcomeModalProps {
  isOpen: boolean;
  onClose: () => void;
  onStartTour: () => void;
}

export function WelcomeModal({ isOpen, onClose, onStartTour }: WelcomeModalProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl" isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <VStack spacing={2} align="center">
            <Text fontSize="3xl">👋</Text>
            <Text>Welcome to OffGridFlow!</Text>
          </VStack>
        </ModalHeader>
        <ModalBody>
          <VStack spacing={4} align="stretch">
            <Text textAlign="center" color={isDark ? 'gray.300' : 'gray.700'}>
              Your comprehensive platform for carbon accounting and compliance management.
            </Text>
            <Box
              p={4}
              bg={isDark ? 'gray.700' : 'gray.50'}
              borderRadius="md"
            >
              <Text fontWeight="bold" mb={2}>
                What you can do:
              </Text>
              <VStack spacing={2} align="start">
                <HStack>
                  <Text>📊</Text>
                  <Text fontSize="sm">Track emissions across Scope 1, 2, and 3</Text>
                </HStack>
                <HStack>
                  <Text>📋</Text>
                  <Text fontSize="sm">Generate compliance reports for CSRD, SEC, CBAM</Text>
                </HStack>
                <HStack>
                  <Text>🔌</Text>
                  <Text fontSize="sm">Connect data sources and automate imports</Text>
                </HStack>
                <HStack>
                  <Text>📈</Text>
                  <Text fontSize="sm">Visualize trends and track reduction targets</Text>
                </HStack>
              </VStack>
            </Box>
          </VStack>
        </ModalBody>
        <ModalFooter>
          <HStack spacing={3} w="full" justify="center">
            <Button variant="ghost" onClick={onClose}>
              Skip for now
            </Button>
            <Button
              colorScheme="brand"
              onClick={() => {
                onClose();
                onStartTour();
              }}
            >
              Take a Tour
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
}

export function useOnboarding() {
  const [hasCompletedOnboarding, setHasCompletedOnboarding] = useState(true);
  const [showWelcome, setShowWelcome] = useState(false);
  const [showTour, setShowTour] = useState(false);

  useEffect(() => {
    // Check if user has completed onboarding
    const completed = localStorage.getItem('onboarding_completed');
    if (!completed) {
      setHasCompletedOnboarding(false);
      setShowWelcome(true);
    }
  }, []);

  const completeOnboarding = () => {
    localStorage.setItem('onboarding_completed', 'true');
    setHasCompletedOnboarding(true);
  };

  const startTour = () => {
    setShowTour(true);
  };

  const completeTour = () => {
    setShowTour(false);
    completeOnboarding();
  };

  return {
    hasCompletedOnboarding,
    showWelcome,
    showTour,
    setShowWelcome,
    startTour,
    completeTour,
  };
}
