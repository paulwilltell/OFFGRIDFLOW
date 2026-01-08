'use client';

import { ToastContainer, toast as toastify, ToastOptions } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { useColorMode } from '@chakra-ui/react';

export function ToastProvider() {
  const { colorMode } = useColorMode();

  return (
    <ToastContainer
      position="top-right"
      autoClose={5000}
      hideProgressBar={false}
      newestOnTop
      closeOnClick
      rtl={false}
      pauseOnFocusLoss
      draggable
      pauseOnHover
      theme={colorMode === 'dark' ? 'dark' : 'light'}
    />
  );
}

export const toast = {
  success: (message: string, options?: ToastOptions) => {
    toastify.success(message, {
      ...options,
      icon: () => '✅',
    });
  },
  error: (message: string, options?: ToastOptions) => {
    toastify.error(message, {
      ...options,
      icon: () => '❌',
    });
  },
  warning: (message: string, options?: ToastOptions) => {
    toastify.warning(message, {
      ...options,
      icon: () => '⚠️',
    });
  },
  info: (message: string, options?: ToastOptions) => {
    toastify.info(message, {
      ...options,
      icon: () => 'ℹ️',
    });
  },
  promise: async <T,>(
    promise: Promise<T>,
    messages: {
      pending: string;
      success: string;
      error: string;
    }
  ) => {
    return toastify.promise(promise, {
      pending: messages.pending,
      success: messages.success,
      error: messages.error,
    });
  },
};
