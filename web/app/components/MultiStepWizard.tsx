'use client';

import {
  Box,
  Button,
  HStack,
  VStack,
  Text,
  Progress,
  Divider,
  useColorMode,
} from '@chakra-ui/react';
import { ReactNode, useState } from 'react';

interface Step {
  title: string;
  description?: string;
  content: ReactNode;
  validate?: () => boolean | Promise<boolean>;
}

interface MultiStepWizardProps {
  steps: Step[];
  onComplete: () => void;
  onCancel?: () => void;
}

export function MultiStepWizard({ steps, onComplete, onCancel }: MultiStepWizardProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isValidating, setIsValidating] = useState(false);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const progress = ((currentStep + 1) / steps.length) * 100;

  const handleNext = async () => {
    const step = steps[currentStep];
    if (step.validate) {
      setIsValidating(true);
      const isValid = await step.validate();
      setIsValidating(false);
      if (!isValid) return;
    }

    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else {
      onComplete();
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  return (
    <Box p={6} bg={isDark ? 'gray.800' : 'white'} borderRadius="lg" boxShadow="lg">
      {/* Progress */}
      <VStack spacing={4} align="stretch" mb={6}>
        <HStack justify="space-between">
          <Text fontSize="sm" fontWeight="medium">
            Step {currentStep + 1} of {steps.length}
          </Text>
          <Text fontSize="sm" color={isDark ? 'gray.400' : 'gray.600'}>
            {Math.round(progress)}% Complete
          </Text>
        </HStack>
        <Progress value={progress} colorScheme="brand" borderRadius="full" />
      </VStack>

      {/* Step Indicators */}
      <HStack spacing={4} mb={6} justify="center" flexWrap="wrap">
        {steps.map((step, index) => (
          <VStack key={index} spacing={1} flex={1} minW="100px">
            <Box
              w="40px"
              h="40px"
              borderRadius="full"
              bg={
                index < currentStep
                  ? 'brand.500'
                  : index === currentStep
                  ? 'brand.500'
                  : isDark
                  ? 'gray.700'
                  : 'gray.200'
              }
              color={index <= currentStep ? 'white' : isDark ? 'gray.400' : 'gray.600'}
              display="flex"
              alignItems="center"
              justifyContent="center"
              fontWeight="bold"
              fontSize="sm"
            >
              {index < currentStep ? '✓' : index + 1}
            </Box>
            <Text
              fontSize="xs"
              fontWeight={index === currentStep ? 'bold' : 'normal'}
              color={index === currentStep ? (isDark ? 'white' : 'black') : isDark ? 'gray.400' : 'gray.600'}
              textAlign="center"
            >
              {step.title}
            </Text>
          </VStack>
        ))}
      </HStack>

      <Divider mb={6} />

      {/* Current Step Content */}
      <Box mb={6}>
        <Text fontSize="xl" fontWeight="bold" mb={2}>
          {steps[currentStep].title}
        </Text>
        {steps[currentStep].description && (
          <Text fontSize="sm" color={isDark ? 'gray.400' : 'gray.600'} mb={4}>
            {steps[currentStep].description}
          </Text>
        )}
        <Box>{steps[currentStep].content}</Box>
      </Box>

      {/* Navigation */}
      <HStack justify="space-between">
        <HStack>
          {currentStep > 0 && (
            <Button onClick={handleBack} variant="outline" isDisabled={isValidating}>
              Back
            </Button>
          )}
          {onCancel && (
            <Button onClick={onCancel} variant="ghost" isDisabled={isValidating}>
              Cancel
            </Button>
          )}
        </HStack>
        <Button onClick={handleNext} colorScheme="brand" isLoading={isValidating}>
          {currentStep < steps.length - 1 ? 'Next' : 'Complete'}
        </Button>
      </HStack>
    </Box>
  );
}
