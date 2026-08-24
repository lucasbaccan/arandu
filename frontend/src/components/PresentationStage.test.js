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

  it('com onFaceClick, clicar num rosto já revelado chama o callback de novo (desrevelar)', async () => {
    const onFaceClick = vi.fn();
    const view = mount({
      pending: [],
      groups: [{ label: 'Go', participants: [ana] }],
      onFaceClick
    });

    const face = view.getByLabelText('Desrevelar resposta de ana@exemplo.com');
    expect(face).not.toBeDisabled();

    await fireEvent.click(face);
    expect(onFaceClick).toHaveBeenCalledWith(ana);
  });

  it('hideZones esconde as zonas e mostra um aviso, mesmo com gente revelada', () => {
    const view = mount({
      pending: [bob],
      groups: [{ label: 'Go', participants: [ana] }],
      hideZones: true
    });

    expect(view.getByText('O organizador escondeu as respostas por enquanto.')).toBeInTheDocument();
    expect(view.queryByText('Go')).not.toBeInTheDocument();
    expect(view.queryByTitle('ana@exemplo.com')).not.toBeInTheDocument();
    // a fila de pendentes continua visível independente do hideZones
    expect(view.getByLabelText('bob@exemplo.com')).toBeInTheDocument();
  });

  it('sem showNames (padrão), não mostra legenda de nome sob os rostos', () => {
    const view = mount({
      pending: [ana],
      groups: [{ label: 'Go', participants: [bob] }]
    });

    expect(view.queryByText('ana@exemplo.com', { selector: '.face-name' })).not.toBeInTheDocument();
    expect(view.queryByText('bob@exemplo.com', { selector: '.face-name' })).not.toBeInTheDocument();
  });

  it('com showNames, mostra o primeiro nome + inicial do sobrenome sob cada rosto (pendente e revelado)', () => {
    const view = mount({
      pending: [{ id: 'p3', name: 'Carla Dias', email: 'carla@exemplo.com', photo: '' }],
      groups: [{ label: 'Go', participants: [{ id: 'p4', name: 'Diego Alves', email: 'diego@exemplo.com', photo: '' }] }],
      showNames: true
    });

    expect(view.getByText('Carla D.')).toBeInTheDocument();
    expect(view.getByText('Diego A.')).toBeInTheDocument();
  });

  it('com showNames, nome de uma palavra só mostra só essa palavra (sem inicial)', () => {
    const view = mount({
      pending: [{ id: 'p3', name: 'Madonna', email: 'madonna@exemplo.com', photo: '' }],
      groups: [],
      showNames: true
    });

    expect(view.getByText('Madonna')).toBeInTheDocument();
  });

  it('rostos pendentes usam estilo tracejado, diferente dos revelados', () => {
    const view = mount({
      pending: [ana],
      groups: [{ label: 'Go', participants: [bob] }]
    });

    const pendingFace = view.getByTitle('ana@exemplo.com');
    const revealedFace = view.getByTitle('bob@exemplo.com');
    expect(pendingFace.classList.contains('pending')).toBe(true);
    expect(revealedFace.classList.contains('pending')).toBe(false);
  });
});
