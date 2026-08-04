import { describe, it, expect, vi } from 'vitest';
import { fireEvent, within } from '@testing-library/dom';
import PresentationStage from './PresentationStage.svelte';

const ana = { id: 'p1', email: 'ana@exemplo.com', photo: '' };
const bob = { id: 'p2', email: 'bob@exemplo.com', photo: '' };

function mount(props) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  new PresentationStage({ target, props });
  return within(target);
}

describe('PresentationStage', () => {
  it('mostra os grupos sempre visíveis, mesmo sem ninguém revelado', () => {
    const view = mount({
      pending: [ana, bob],
      groups: [
        { label: 'Go', participants: [] },
        { label: 'JS', participants: [] }
      ]
    });

    expect(view.getByText('Go')).toBeInTheDocument();
    expect(view.getByText('JS')).toBeInTheDocument();
    expect(view.getAllByText('0')).toHaveLength(2);
  });

  it('sem onFaceClick, os rostos pendentes ficam desabilitados (somente leitura)', () => {
    const view = mount({
      pending: [ana],
      groups: []
    });

    const face = view.getByTitle('ana@exemplo.com');
    expect(face.tagName).toBe('BUTTON');
    expect(face).toBeDisabled();
  });

  it('com onFaceClick, os rostos pendentes ficam clicáveis e chamam o callback', async () => {
    const onFaceClick = vi.fn();
    const view = mount({
      pending: [ana],
      groups: [{ label: 'Go', participants: [] }],
      onFaceClick
    });

    const face = view.getByLabelText('Revelar resposta de ana@exemplo.com');
    expect(face).not.toBeDisabled();

    await fireEvent.click(face);
    expect(onFaceClick).toHaveBeenCalledWith(ana);
  });

  it('mostra "Ninguém respondeu ainda" quando não há ninguém, e "Todas as respostas foram reveladas" quando todo mundo já foi revelado', () => {
    const emptyView = mount({ pending: [], groups: [{ label: 'Go', participants: [] }] });
    expect(emptyView.getByText('Ninguém respondeu ainda.')).toBeInTheDocument();

    const allRevealedView = mount({
      pending: [],
      groups: [{ label: 'Go', participants: [ana] }]
    });
    expect(allRevealedView.getByText('Todas as respostas foram reveladas.')).toBeInTheDocument();
  });

  it('renderiza participantes revelados dentro do grupo certo, com contagem', () => {
    const view = mount({
      pending: [bob],
      groups: [
        { label: 'Go', participants: [ana] },
        { label: 'JS', participants: [] }
      ]
    });

    const goZone = view.getByText('Go').closest('.zone');
    expect(within(goZone).getByTitle('ana@exemplo.com')).toBeInTheDocument();
    expect(within(goZone).getByText('1')).toBeInTheDocument();

    const jsZone = view.getByText('JS').closest('.zone');
    expect(within(jsZone).getByText('0')).toBeInTheDocument();

    expect(view.getByLabelText('bob@exemplo.com')).toBeInTheDocument();
  });
});
