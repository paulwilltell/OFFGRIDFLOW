'use client';

import { useEffect, useRef } from 'react';

type RichTextInputProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  disabled?: boolean;
};

const toolbarButtons: Array<{ label: string; command: string }> = [
  { label: 'B', command: 'bold' },
  { label: 'I', command: 'italic' },
  { label: '•', command: 'insertUnorderedList' },
];

export function RichTextInput({
  value,
  onChange,
  placeholder,
  disabled = false,
}: RichTextInputProps) {
  const editorRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!editorRef.current) {
      return;
    }

    if (editorRef.current.innerHTML !== value) {
      editorRef.current.innerHTML = value;
    }
  }, [value]);

  const applyCommand = (command: string) => {
    if (disabled) {
      return;
    }

    editorRef.current?.focus();
    document.execCommand(command);
    onChange(editorRef.current?.innerHTML ?? '');
  };

  return (
    <div className="rounded-2xl border border-slate-700/70 bg-slate-950/80">
      <div className="flex items-center gap-2 border-b border-slate-800/80 px-3 py-2">
        {toolbarButtons.map((button) => (
          <button
            key={button.command}
            type="button"
            className="rounded-md border border-slate-700 px-2 py-1 text-xs font-semibold text-slate-200 transition hover:border-emerald-400 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
            onClick={() => applyCommand(button.command)}
            disabled={disabled}
          >
            {button.label}
          </button>
        ))}
      </div>
      <div className="relative">
        {!value && (
          <div className="pointer-events-none absolute left-3 top-3 right-3 text-sm leading-6 text-slate-500">
            {placeholder}
          </div>
        )}
        <div
          ref={editorRef}
          contentEditable={!disabled}
          suppressContentEditableWarning
          onInput={() => onChange(editorRef.current?.innerHTML ?? '')}
          className="min-h-[144px] px-3 py-3 text-sm leading-6 text-slate-100 outline-none"
        />
      </div>
    </div>
  );
}
