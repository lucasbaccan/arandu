package main

// Conteúdo fake usado para popular o banco em modo demo. Tudo em pt-BR pra
// combinar com o resto do produto.

// eventStatuses espelha os status válidos de internal/api/eventos.go
// (constantes não exportadas lá, por isso repetidas aqui).
var eventStatuses = []string{"PREPARATION", "OPEN_FOR_ANSWERS", "CLOSED_FOR_ANSWERS"}

var eventTitles = []string{
	"Sprint Review - Squad Foguete",
	"Onboarding de Novos Talentos",
	"Culture Day - Tribo Norte",
	"Kickoff do Projeto Aurora",
	"Café com o Time - Squad Estrela",
	"Retrospectiva do Trimestre",
	"Integração de Estagiários 2026",
	"Workshop de Liderança",
	"All Hands - Arandu",
	"Confraternização de Fim de Ano",
}

// open15Question é a pergunta aberta que precisa acumular 15 respostas
// distintas (uma "opção" de texto livre por participante do pool abaixo).
const open15Question = "Qual filme ou série marcou a sua infância ou adolescência?"

var open15Pool = []string{
	"Matrix",
	"Friends",
	"Squid Game",
	"Harry Potter",
	"O Senhor dos Anéis",
	"Breaking Bad",
	"Toy Story",
	"Chaves",
	"A Casa de Papel",
	"Stranger Things",
	"Vingadores: Ultimato",
	"Star Wars",
	"Grey's Anatomy",
	"O Auto da Compadecida",
	"A Fantástica Fábrica de Chocolate",
}

// openGroupQuestion é a pergunta aberta usada pra demonstrar o agrupamento
// de respostas repetidas com capitalização diferente (ver normalizeOpenText
// em internal/api/live.go: só a primeira letra é normalizada).
const openGroupQuestion = "Qual sua linguagem de programação favorita?"

// groupWordPool fica em minúsculas de propósito: seedAnswers decide, por
// participante, se capitaliza a primeira letra ou não, criando duplicatas
// que devem ser agrupadas (mesma palavra, primeira letra em caixas
// diferentes).
var groupWordPool = []string{"python", "javascript", "go", "java", "ruby", "c#"}

var genericOpenQuestions = []string{
	"O que você mais gosta de fazer no seu tempo livre?",
	"Qual foi a sua maior conquista neste ano?",
	"Se pudesse aprender uma nova habilidade instantaneamente, qual seria?",
	"O que te motiva a vir trabalhar todos os dias?",
	"Qual conselho você daria para quem está começando agora?",
	"O que você faria se tivesse um dia de folga surpresa?",
}

var genericAnswerPool = []string{
	"Aprender algo novo todo dia.",
	"Viajar mais este ano.",
	"Passar mais tempo com a família.",
	"Ler pelo menos um livro por mês.",
	"Fazer exercícios com mais frequência.",
	"Aprender a tocar violão.",
	"Cozinhar pratos novos.",
	"Meditar todas as manhãs.",
	"Assistir mais séries.",
	"Jogar videogame nos fins de semana.",
	"Correr uma maratona.",
	"Aprender outro idioma.",
}

// longOpenTitle e longGroupTitle são usados pra garantir 2 perguntas com
// título bem grande por evento (uma aberta, uma de múltipla escolha).
const longOpenTitle = "Pensando em tudo que vivemos juntos nos últimos meses, entre entregas apertadas, aprendizados inesperados, cafés reforçados e aquelas conversas de corredor que resolvem mais problema do que qualquer reunião marcada, qual foi o momento que você mais guardaria com carinho na memória deste time?"

const longGroupTitle = "Se você pudesse escolher exatamente uma única palavra ou expressão para representar o espírito deste time durante essa fase tão intensa de entregas, reuniões maratona, decisões de última hora e xingamentos carinhosos no grupo do chat, qual das opções abaixo combina mais com a vibe geral do squad neste momento?"

type groupQuestion struct {
	Title   string
	Options []string
}

var groupQuestionPool = []groupQuestion{
	{"Qual seu prato favorito?", []string{"Pizza", "Sushi", "Feijoada", "Churrasco", "Hambúrguer"}},
	{"Qual estação do ano você prefere?", []string{"Verão", "Outono", "Inverno", "Primavera"}},
	{"Qual super-herói você seria?", []string{"Homem de Ferro", "Mulher Maravilha", "Batman", "Flash", "Hulk"}},
	{"Como você prefere trabalhar?", []string{"Home office", "Presencial", "Híbrido"}},
	{"Qual seu animal de estimação favorito?", []string{"Cachorro", "Gato", "Pássaro", "Nenhum"}},
	{"Qual seu esporte favorito?", []string{"Futebol", "Vôlei", "Natação", "Corrida", "Nenhum"}},
	{"Qual seu estilo musical favorito?", []string{"Rock", "Pop", "Sertanejo", "MPB", "Eletrônica"}},
	{"Qual período do dia você rende mais?", []string{"Manhã", "Tarde", "Noite", "Madrugada"}},
}

type participantSeed struct {
	Name  string
	Local string // parte local do e-mail (antes do @), sem acentos.
}

var namePool = []participantSeed{
	{"Ana Souza", "ana.souza"},
	{"Bruno Lima", "bruno.lima"},
	{"Carla Mendes", "carla.mendes"},
	{"Diego Alves", "diego.alves"},
	{"Elisa Ramos", "elisa.ramos"},
	{"Fábio Nunes", "fabio.nunes"},
	{"Gabriela Rocha", "gabriela.rocha"},
	{"Hugo Teixeira", "hugo.teixeira"},
	{"Isabela Martins", "isabela.martins"},
	{"João Pereira", "joao.pereira"},
	{"Karina Dias", "karina.dias"},
	{"Leonardo Cardoso", "leonardo.cardoso"},
	{"Mariana Costa", "mariana.costa"},
	{"Nicolas Freitas", "nicolas.freitas"},
	{"Olivia Barros", "olivia.barros"},
	{"Pedro Henrique Santos", "pedro.santos"},
	{"Quesia Fernandes", "quesia.fernandes"},
	{"Rafael Gomes", "rafael.gomes"},
	{"Sofia Araújo", "sofia.araujo"},
	{"Thiago Moreira", "thiago.moreira"},
	{"Ursula Pinto", "ursula.pinto"},
	{"Vinícius Correia", "vinicius.correia"},
	{"Wesley Batista", "wesley.batista"},
	{"Ximena Duarte", "ximena.duarte"},
	{"Yasmin Ribeiro", "yasmin.ribeiro"},
	{"Zeca Monteiro", "zeca.monteiro"},
	{"Aline Farias", "aline.farias"},
	{"Bianca Castro", "bianca.castro"},
	{"Caio Vasconcelos", "caio.vasconcelos"},
	{"Débora Azevedo", "debora.azevedo"},
}
