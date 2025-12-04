'use client';

import {
  Box,
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  useColorMode,
  HStack,
  Text,
} from '@chakra-ui/react';
import { useTranslation } from 'react-i18next';

export function ThemeSwitcher() {
  const { colorMode, toggleColorMode } = useColorMode();

  return (
    <IconButton
      aria-label={`Switch to ${colorMode === 'light' ? 'dark' : 'light'} mode`}
      icon={<span style={{ fontSize: '20px' }}>{colorMode === 'light' ? '🌙' : '☀️'}</span>}
      onClick={toggleColorMode}
      variant="ghost"
      size="md"
    />
  );
}

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const languages = [
    { code: 'en', label: 'English', flag: '🇺🇸' },
    { code: 'es', label: 'Español', flag: '🇪🇸' },
    { code: 'de', label: 'Deutsch', flag: '🇩🇪' },
    { code: 'fr', label: 'Français', flag: '🇫🇷' },
  ];

  const currentLanguage = languages.find((lang) => lang.code === i18n.language) || languages[0];

  return (
    <Menu>
      <MenuButton
        as={IconButton}
        aria-label="Change language"
        icon={<span style={{ fontSize: '20px' }}>{currentLanguage.flag}</span>}
        variant="ghost"
        size="md"
      />
      <MenuList>
        {languages.map((lang) => (
          <MenuItem
            key={lang.code}
            onClick={() => i18n.changeLanguage(lang.code)}
            fontWeight={i18n.language === lang.code ? 'bold' : 'normal'}
          >
            <HStack>
              <span>{lang.flag}</span>
              <Text>{lang.label}</Text>
            </HStack>
          </MenuItem>
        ))}
      </MenuList>
    </Menu>
  );
}

export function ThemeAndLanguageControls() {
  return (
    <HStack spacing={2}>
      <ThemeSwitcher />
      <LanguageSwitcher />
    </HStack>
  );
}
