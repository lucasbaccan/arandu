import { describe, it, expect } from 'vitest';
import { formatDate, formatDateTime } from './formatDate.js';

describe('formatDate', () => {
  it('formata no padrão DD/MM/YYYY', () => {
    expect(formatDate('2026-08-02T12:00:00Z')).toBe('02/08/2026');
  });

  it('não desloca o dia por fuso horário (UTC-3: 00hZ seria 21h do dia anterior)', () => {
    expect(formatDate('2026-08-02T00:00:00Z')).toBe('02/08/2026');
  });

  it('formata datas com dia e mês de um dígito com zero à esquerda', () => {
    expect(formatDate('2026-01-05T10:30:00Z')).toBe('05/01/2026');
  });
});

describe('formatDateTime', () => {
  it('formata no padrão DD/MM/YYYY HH:mm', () => {
    expect(formatDateTime('2026-08-02T14:05:00Z')).toBe('02/08/2026 14:05');
  });

  it('preenche hora e minuto de um dígito com zero à esquerda', () => {
    expect(formatDateTime('2026-08-02T09:03:00Z')).toBe('02/08/2026 09:03');
  });
});
