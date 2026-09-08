## [1.1.1](https://github.com/lucasbaccan/arandu/compare/v1.1.0...v1.1.1) (2026-09-08)


### Bug Fixes

* **ui:** injeta o commit real no build do Docker e ajusta o texto da home ([#9](https://github.com/lucasbaccan/arandu/issues/9)) ([88ddf6e](https://github.com/lucasbaccan/arandu/commit/88ddf6e69c8d5f6a559f0aa5d0b084f3e66db57d))

# [1.1.0](https://github.com/lucasbaccan/arandu/compare/v1.0.0...v1.1.0) (2026-09-08)


### Features

* **conta:** trocar senha sem a senha atual e olho de mostrar senha ([#8](https://github.com/lucasbaccan/arandu/issues/8)) ([8065f47](https://github.com/lucasbaccan/arandu/commit/8065f477e7c3f8778384d1621d10fd260c9fb4ae))

# 1.0.0 (2026-09-08)


* feat!: versão estável 1.0.0 — apresentação ao vivo resiliente ([6047b18](https://github.com/lucasbaccan/arandu/commit/6047b1847bc480abbc33708bb7e003c34b996c27))


### Bug Fixes

* acessibilidade do Card clicável, volta ao início no login/cadastro e altura no mobile ([5e1dd45](https://github.com/lucasbaccan/arandu/commit/5e1dd45159a73991302df0e8f2fa9b53efccaebc)), closes [#app](https://github.com/lucasbaccan/arandu/issues/app)
* alinha validacao de e-mail entre frontend e backend ([ee7e71d](https://github.com/lucasbaccan/arandu/commit/ee7e71d796fa0d6c702fbb65f02633945e289a77))
* **back:** cookie de sessão decide Secure/SameSite por requisição ([a70577b](https://github.com/lucasbaccan/arandu/commit/a70577be9812e20799e6238defc661a3abab3c6f))
* **back:** CORS wildcard com esquema (https://*.vercel.app) ([e5c67d7](https://github.com/lucasbaccan/arandu/commit/e5c67d7cb1261d2f62888a712ab727581e71fe4b))
* **front:** botões de navegação do dock de /stage iguais (estilo secondary) ([52597e9](https://github.com/lucasbaccan/arandu/commit/52597e9024df105825e18c70b15670c2af5ad553))
* **front:** dist dentro do frontend no build da Vercel (VERCEL=1) ([530b9ec](https://github.com/lucasbaccan/arandu/commit/530b9ec2f25d434476f526b3aff8942f830f468f))
* gera edit_token para participantes legados sem token ([bc8ba72](https://github.com/lucasbaccan/arandu/commit/bc8ba72e622babe9fa1f1f30ac0bc79724e06b72))
* remove do tab as perguntas fora da tela no carrossel de respostas ([0408758](https://github.com/lucasbaccan/arandu/commit/04087582174d42997edf266dc1a6273ea2c63cb4))
* **stage:** animações do placar não vazam mais sobre o trilho e o resto da tela ([46b186c](https://github.com/lucasbaccan/arandu/commit/46b186c552ed8c6bb50ce79643dde975f4a18d6c))


### Features

* adiciona confirmacao antes de remover uma pergunta ([97353b3](https://github.com/lucasbaccan/arandu/commit/97353b3ab30199057affb9b7083e57c1abb83f66))
* adiciona fluxo público de participação e envio de respostas ([0da1385](https://github.com/lucasbaccan/arandu/commit/0da1385154ca164d48cf4d6db725a8d050868f1c))
* adiciona painel do organizador para listar e editar respostas ([445831a](https://github.com/lucasbaccan/arandu/commit/445831a069f0516a18344708bcc9be4da58a6caf))
* adiciona preview da apresentacao com revelacao animada dos rostos ([45235b5](https://github.com/lucasbaccan/arandu/commit/45235b5981f772421dff82d69f0df97b172a6f1f))
* adiciona status de respostas abertas/fechadas ao evento ([15017ab](https://github.com/lucasbaccan/arandu/commit/15017ab6c371f44cf6753caa254bd092f8ed9cde))
* adiciona tipo de pergunta de resposta aberta (OPEN_TEXT) ([5c7fff6](https://github.com/lucasbaccan/arandu/commit/5c7fff68b6362f0a2d55b985d3f08343ca336465))
* alternancia de respostas abertas/fechadas na edicao de evento ([d93feea](https://github.com/lucasbaccan/arandu/commit/d93feea35a7563617de552e40e4dcf751cca2469))
* apresentacao publica ao vivo via SSE, com entrada por PIN + e-mail ([9d39589](https://github.com/lucasbaccan/arandu/commit/9d395893c9cd80a397d83b6820f35554101f9b24))
* autenticacao com telas de login e cadastro ([6314704](https://github.com/lucasbaccan/arandu/commit/6314704e10e10947fb402eed48cdcb9c58cee3a7))
* **back:** CORS com wildcard de sufixo (*.vercel.app) para previews ([35fbdf9](https://github.com/lucasbaccan/arandu/commit/35fbdf941672542f4e30955310344ea72f523823))
* **back:** CORS configurável (CORS_ORIGINS) e cookie SameSite cross-site ([e1ffaf8](https://github.com/lucasbaccan/arandu/commit/e1ffaf89d8f43658389c16dad76c070cfa37b271))
* **back:** fotos dos participantes servidas como arquivos, fora do payload ([d364df2](https://github.com/lucasbaccan/arandu/commit/d364df25d0191e3b14aa732f3891cf6c270d7117))
* **back:** permite desabilitar edição de respostas por evento ([eca311b](https://github.com/lucasbaccan/arandu/commit/eca311bd850cbb2958fe1d00161e9944bcad1ba6))
* **back:** Q&A ao vivo — participante remove a própria pergunta ([d3e5925](https://github.com/lucasbaccan/arandu/commit/d3e592595576f888f440a50b3dc059da2db1e493))
* botao de copiar PIN e link de participacao ([5c920f2](https://github.com/lucasbaccan/arandu/commit/5c920f2ee9de5762b43a34d8c4cda683c3886f74))
* **conta:** trocar senha logado (senha atual + nova) ([3a3e922](https://github.com/lucasbaccan/arandu/commit/3a3e92276ed126e9fc1a56af0c3504636fb639be))
* **design:** evolui o design system com nova paleta e componentes ([4ea01d2](https://github.com/lucasbaccan/arandu/commit/4ea01d2a8e9abbf8087b6525aae8181a1cc53efd))
* dev flow com Go servindo o frontend, particulas globais e fixes ([6179f7e](https://github.com/lucasbaccan/arandu/commit/6179f7e1361858ab1526679621e88aeea30822d6))
* edicao de evento, toasts, setup interativo e melhorias ([ff50f37](https://github.com/lucasbaccan/arandu/commit/ff50f37403eaf29c21e3a08e1bdd03479726d8f0))
* **front:** mapa do site (/mapa-do-site) e tela de debug (/debug) ([73db7d4](https://github.com/lucasbaccan/arandu/commit/73db7d4feb5a41424833e2e0a40dc0f61466b13f))
* **front:** Material Symbols Rounded, tokens de contraste AA e alvos de toque ([bacb406](https://github.com/lucasbaccan/arandu/commit/bacb406bef76117e6b5978d1c51ff234c1e3c221))
* **front:** novo design system (TopBar/CrumbBar/PublicShell + tokens) ([14209c8](https://github.com/lucasbaccan/arandu/commit/14209c8bf66f1b0034dc2df63dc79462b107f473))
* **front:** tela da plateia unificada com reações e Q&A ([a4d5412](https://github.com/lucasbaccan/arandu/commit/a4d54128641cb1793d8a0dd15b2cd03561398dbe))
* **front:** trata sessão expirada de forma centralizada ([b2450fd](https://github.com/lucasbaccan/arandu/commit/b2450fdc626f6dd5197af3280742dc6436a7a6c7))
* **front:** VITE_BACKEND_URL para rodar o frontend separado do backend (Vercel) ([be66d77](https://github.com/lucasbaccan/arandu/commit/be66d776c8bc408278f2763f4356c5e9aeb98f5d))
* **icones:** página de decisão de ícones com galeria Auto/Smart ([5353b85](https://github.com/lucasbaccan/arandu/commit/5353b85f6a210195643acd386d579937c59885ea))
* identidade visual Porandu com logo proprio ([46469a1](https://github.com/lucasbaccan/arandu/commit/46469a121b96c580593e6250e44d5d173a7f9f35))
* link de edicao de respostas e edicao de foto pelo participante ([89a3576](https://github.com/lucasbaccan/arandu/commit/89a3576d2ba97ef76e045b314fec3d0bac3e8234))
* **live:** janela de apresentação separada, toggle de nomes e reset-all na UI ([2949199](https://github.com/lucasbaccan/arandu/commit/2949199b15d146feff0aca35988c6b5118257e4f))
* **live:** liga a UI ao desrevelar e ao ocultar respostas ([3266658](https://github.com/lucasbaccan/arandu/commit/3266658bba97051de3c46aa90214b505a2964647))
* **live:** permite desrevelar participante e ocultar respostas na apresentação ([9ff5ef8](https://github.com/lucasbaccan/arandu/commit/9ff5ef8e05d95aa229498b94233e051ed516b822))
* **live:** persiste estado ao vivo no banco, esconde nomes, reset-all e modo apresentação ([5350679](https://github.com/lucasbaccan/arandu/commit/5350679ee9451549dae9b1752286de48238d8232))
* modos de visualização no palco, excluir evento e imagem Docker ([b721efd](https://github.com/lucasbaccan/arandu/commit/b721efde5ffa0ec7bad40033ffbfb90af98f82fb))
* mostra data e hora de quando cada pessoa respondeu ([23b6278](https://github.com/lucasbaccan/arandu/commit/23b62787ffe87eac193f40e3ca67cfd88d9247d2))
* normaliza PIN do evento para maiúsculas ([9ce0337](https://github.com/lucasbaccan/arandu/commit/9ce03378e0efd393b2e5f70230d66ade13098d3e))
* painel de respostas do organizador com abas e corrige drag-and-drop ([f0c77d7](https://github.com/lucasbaccan/arandu/commit/f0c77d74afd95102362ef1fc7d912a758f936d4f))
* **palco:** modos de visualização Compacto/Amplo com ícones Material e dock mobile ([54f5fe6](https://github.com/lucasbaccan/arandu/commit/54f5fe60c1cacb04c0e6daf0318caa38291f75f0))
* **participants:** guarda e exibe o nome do participante na apresentação ([117d83e](https://github.com/lucasbaccan/arandu/commit/117d83e9f104edb744a33b1ef5a1556672fdf43b))
* pergunta de resposta aberta (OPEN_TEXT) no formulario de perguntas ([7fc226b](https://github.com/lucasbaccan/arandu/commit/7fc226b7b686bcd8be57e2c0e32df114328456a0))
* perguntas na edicao de evento com reordenacao e edicao ([29c8b08](https://github.com/lucasbaccan/arandu/commit/29c8b08a57926bb8c76c19114f32eb79a49209ab))
* **present:** densidade do placar vira estado ao vivo via SSE ([c4b80fd](https://github.com/lucasbaccan/arandu/commit/c4b80fd03a1f0d670b1e89c729751f8879cc4bfa))
* reações, mensagens e Q&A ao vivo entre palco e plateia ([9f6d077](https://github.com/lucasbaccan/arandu/commit/9f6d077b3cb9cb982aa645bb5c41b922e25da56a))
* rebrand visual da marca para Arandu ([c28df42](https://github.com/lucasbaccan/arandu/commit/c28df4218b6401d77d06cb868b6b763afcd663fe))
* redesenha dashboard, edicao de evento e apresentacao ao vivo ([937de52](https://github.com/lucasbaccan/arandu/commit/937de526c1db4a35955d6f6e17462a21f5939a98))
* **seed:** make seed popula o banco com dados de demonstração ([86ceb5f](https://github.com/lucasbaccan/arandu/commit/86ceb5f89eb0e0aeef4cde62a8e5deefa82e623b))
* simplifica edicao do PIN e mostra quantidade de respostas ([8d30455](https://github.com/lucasbaccan/arandu/commit/8d304552855bd118175b313bbefe7a7106353304))
* **stage:** dicas de hover no placar com popup interativo de revelação ([7950a5e](https://github.com/lucasbaccan/arandu/commit/7950a5e17f4122bdd67e32be744905616959ac55))
* **stage:** layout compact em telas pequenas e contraste das cores de zona ([471bdd0](https://github.com/lucasbaccan/arandu/commit/471bdd0fe7b285f42cef7baa92421a47f69a32fb))
* **stage:** placar de respostas cabe na tela sem cortar opções ([ac77173](https://github.com/lucasbaccan/arandu/commit/ac771730cae758510a795362a3099300bed70751))
* tela de criacao de eventos e dashboard com lista ([9f04b08](https://github.com/lucasbaccan/arandu/commit/9f04b0893961c7783bb174e0e0a18a83f39c7b61))
* tela de resposta do participante ([bf0b0bd](https://github.com/lucasbaccan/arandu/commit/bf0b0bde6cb62eb6dd7c5164ebc06bec6215fd38))
* tela publica /live/:id e sincroniza o admin com ela em tempo real ([7406784](https://github.com/lucasbaccan/arandu/commit/74067849856cf700b1f916e9dac0d94422d884a7))
* **ui:** ajustes mobile e alvos de toque nos componentes ([4a1f338](https://github.com/lucasbaccan/arandu/commit/4a1f338ed2ea9c09557a25cea422143274a05453))
* variaveis HOST/PORT/FRONTEND_PORT no Makefile ([ca6a2f4](https://github.com/lucasbaccan/arandu/commit/ca6a2f4135529baa3e7328503703e77dcdc8ade9))
* **views:** telas de organizador — mobile e contraste ([e24a6a2](https://github.com/lucasbaccan/arandu/commit/e24a6a2d9cc3605d0db1405daea723f6b5913bd0))
* **views:** telas públicas — mobile, contraste e acessibilidade ([31976d2](https://github.com/lucasbaccan/arandu/commit/31976d26446c636441df9751333cecd79192050c))


### Performance Improvements

* **back:** navegador cacheia fotos por 1h antes de revalidar ([18052ab](https://github.com/lucasbaccan/arandu/commit/18052abe23e678750919819ff6b6f834b479534d))


### BREAKING CHANGES

* primeira versão estável (0.x -> 1.0.0)
