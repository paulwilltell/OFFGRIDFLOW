'use client';

import { useState } from 'react';

interface ColorTheme {
  primary: string;
  secondary: string;
  background: string;
  text: string;
}

const PRESET_THEMES: Record<string, ColorTheme> = {
  default: {
    primary: '#8aa9ff',
    secondary: '#1d2940',
    background: '#0a0f1e',
    text: '#ffffff',
  },
  green: {
    primary: '#10b981',
    secondary: '#064e3b',
    background: '#022c22',
    text: '#ffffff',
  },
  purple: {
    primary: '#a78bfa',
    secondary: '#4c1d95',
    background: '#1e1b4b',
    text: '#ffffff',
  },
  orange: {
    primary: '#fb923c',
    secondary: '#7c2d12',
    background: '#431407',
    text: '#ffffff',
  },
};

interface ThemeCustomizerProps {
  onThemeChange?: (theme: ColorTheme) => void;
}

export default function ThemeCustomizer({ onThemeChange }: ThemeCustomizerProps) {
  const [selectedPreset, setSelectedPreset] = useState<string>('default');
  const [customColors, setCustomColors] = useState<ColorTheme>(PRESET_THEMES.default);
  const [customMode, setCustomMode] = useState(false);

  const applyPreset = (presetName: string) => {
    const theme = PRESET_THEMES[presetName];
    setSelectedPreset(presetName);
    setCustomColors(theme);
    setCustomMode(false);
    
    if (onThemeChange) {
      onThemeChange(theme);
    }
  };

  const updateCustomColor = (key: keyof ColorTheme, value: string) => {
    const newTheme = { ...customColors, [key]: value };
    setCustomColors(newTheme);
    setCustomMode(true);
    
    if (onThemeChange) {
      onThemeChange(newTheme);
    }
  };

  const exportTheme = () => {
    const css = `
:root {
  --primary-color: ${customColors.primary};
  --secondary-color: ${customColors.secondary};
  --background-color: ${customColors.background};
  --text-color: ${customColors.text};
}
    `.trim();

    const blob = new Blob([css], { type: 'text/css' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'theme.css';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div style={{ padding: '1.5rem', background: '#0f172a', borderRadius: '12px', border: '1px solid #1d2940' }}>
      <h3 style={{ fontSize: '1.1rem', marginBottom: '1rem', color: '#8aa9ff' }}>Theme Customizer</h3>

      {/* Preset themes */}
      <div style={{ marginBottom: '1.5rem' }}>
        <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.75rem', color: '#fff' }}>
          Preset Themes
        </div>
        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
          {Object.keys(PRESET_THEMES).map((presetName) => {
            const theme = PRESET_THEMES[presetName];
            return (
              <button
                key={presetName}
                onClick={() => applyPreset(presetName)}
                style={{
                  padding: '0.75rem 1rem',
                  background: selectedPreset === presetName ? theme.primary : '#1d2940',
                  color: selectedPreset === presetName ? theme.background : '#fff',
                  border: selectedPreset === presetName ? `2px solid ${theme.primary}` : '1px solid #374151',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  fontSize: '0.85rem',
                  textTransform: 'capitalize',
                  fontWeight: selectedPreset === presetName ? 600 : 400,
                }}
              >
                {presetName}
              </button>
            );
          })}
        </div>
      </div>

      {/* Custom colors */}
      <div style={{ marginBottom: '1.5rem' }}>
        <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.75rem', color: '#fff' }}>
          Custom Colors
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
          {Object.entries(customColors).map(([key, value]) => (
            <div key={key}>
              <label
                htmlFor={`color-${key}`}
                style={{
                  display: 'block',
                  fontSize: '0.8rem',
                  color: '#888',
                  marginBottom: '0.25rem',
                  textTransform: 'capitalize',
                }}
              >
                {key}
              </label>
              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                <input
                  id={`color-${key}`}
                  type="color"
                  value={value}
                  onChange={(e) => updateCustomColor(key as keyof ColorTheme, e.target.value)}
                  style={{
                    width: '50px',
                    height: '40px',
                    border: '1px solid #374151',
                    borderRadius: '6px',
                    cursor: 'pointer',
                  }}
                />
                <input
                  type="text"
                  value={value}
                  onChange={(e) => updateCustomColor(key as keyof ColorTheme, e.target.value)}
                  style={{
                    flex: 1,
                    padding: '0.5rem',
                    background: '#1d2940',
                    color: '#fff',
                    border: '1px solid #374151',
                    borderRadius: '6px',
                    fontSize: '0.85rem',
                    fontFamily: 'monospace',
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Preview */}
      <div style={{ marginBottom: '1.5rem' }}>
        <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.75rem', color: '#fff' }}>
          Preview
        </div>
        <div
          style={{
            padding: '1.5rem',
            background: customColors.background,
            color: customColors.text,
            borderRadius: '8px',
            border: `2px solid ${customColors.secondary}`,
          }}
        >
          <div style={{ fontSize: '1.2rem', fontWeight: 700, marginBottom: '0.5rem', color: customColors.primary }}>
            Sample Heading
          </div>
          <p style={{ marginBottom: '1rem', fontSize: '0.9rem' }}>
            This is how your content will look with the selected theme colors.
          </p>
          <button
            style={{
              padding: '0.5rem 1rem',
              background: customColors.primary,
              color: customColors.background,
              border: 'none',
              borderRadius: '6px',
              fontSize: '0.85rem',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Sample Button
          </button>
        </div>
      </div>

      {/* Export */}
      <button
        onClick={exportTheme}
        style={{
          width: '100%',
          padding: '0.75rem',
          background: '#8aa9ff',
          color: '#0a0f1e',
          border: 'none',
          borderRadius: '8px',
          fontSize: '0.9rem',
          fontWeight: 600,
          cursor: 'pointer',
        }}
      >
        💾 Export Theme CSS
      </button>
    </div>
  );
}
