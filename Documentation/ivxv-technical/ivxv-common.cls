% ivxv-common.cls

% Class definition
\ProvidesClass{ivxv-common}

% Class options for handling documents in different languages
\newcommand{\@IVXVStringConfidential}{Konfidentsiaalne}
\newcommand{\@IVXVStringDate}{Kuup\"aev:}
\newcommand{\@IVXVStringDocumentType}{Tehniline dokument}
\newcommand{\@IVXVStringVersion}{Versioon}
\newcommand{\@IVXVStringPages}{lk}
\newcommand{\@IVXVStringDocument}{Dok}
\newcommand{\@IVXVStringBackofTitlePage}{\copyright\ RVT, \the\year}
\newcommand{\@IVXVStringCopyrightFooter}{\copyright\ RVT, \the\year}

\DeclareOption{et-lang}{%
	\renewcommand{\@IVXVStringConfidential}{Konfidentsiaalne}%
	\renewcommand{\@IVXVStringDate}{Kuup\"aev:}%
	\renewcommand{\@IVXVStringDocumentType}{Tehniline dokument}%
	\renewcommand{\@IVXVStringVersion}{Versioon}%
	\renewcommand{\@IVXVStringPages}{lk}%
	\renewcommand{\@IVXVStringDocument}{Dok}%
}
\DeclareOption{en-lang}{%
	\renewcommand{\@IVXVStringConfidential}{Confidential}%
	\renewcommand{\@IVXVStringDate}{Date:}%
	\renewcommand{\@IVXVStringDocumentType}{Technical document}%
	\renewcommand{\@IVXVStringVersion}{Version}%
	\renewcommand{\@IVXVStringPages}{pages}%
	\renewcommand{\@IVXVStringDocument}{Doc.}%
}

\ProcessOptions\relax


% Base class, which we will modify
% Add oldfontcommands for newer versions of memoir.
%\LoadClass[a4paper, twoside, 12pt, article]{memoir}
\LoadClass[a4paper, twoside, 12pt, article, oldfontcommands]{memoir}

\usepackage{cmap}

% Make the figure and table captions separated by full stops rather than colons.
\usepackage{caption}
\captionsetup[table]{labelsep=period}
\captionsetup[figure]{labelsep=period}

% margins
\usepackage[top=1in,bottom=1in,left=1in,right=1in,pdftex]{geometry}

% For figuring out, how many pages there is in the document
\usepackage{lastpage}

% We want to do later MakeUppercase for the section heads and when \label{} is included in the section head, then it will break everything.
% Textcase package provides more intelligent MakeUppercase command
\usepackage[overload]{textcase}

% utf8 fileformat
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
% Times New Roman font for body text
% Arial font for section headlines. The font can be installed
% with ftp://tug.org/tex/getnonfreefonts/install-getnonfreefonts utility
%\usepackage[scaled]{uarial}
\renewcommand*{\familydefault}{\sfdefault}

\newcommand{\IVXVArial}{\fontfamily{phv} \selectfont}

% Commands for setting the document type, number, version, etc.
\newcommand{\IVXVDefineDocumentType}[1]{\def\@IVXVStringDocumentType{#1}}
\newcommand{\IVXVDefineDocumentNumber}[1]{\def\@IVXVDocumentNumber{#1}}
\newcommand{\IVXVDefineDocumentVersion}[1]{\def\@IVXVDocumentVersion{#1}}
\newcommand{\IVXVDefineDocumentSecurity}[1]{\renewcommand{\@IVXVStringConfidential}{#1}}
\newcommand{\IVXVDefineAcknowledgement}[1]{\def\@IVXVAcknowledgement{#1}}
\newcommand{\IVXVDefineCopyrightFooter}[1]{\renewcommand{\@IVXVStringCopyrightFooter}{#1}}
\newcommand{\IVXVDefineBackofTitlePage}[1]{\renewcommand{\@IVXVStringBackofTitlePage}{#1}}

% .. and initialize them to empty
\IVXVDefineDocumentNumber{}
\global\let\@IVXVDocumentNumber\@empty
\IVXVDefineDocumentVersion{}
\global\let\@IVXVDocumentVersion\@empty
\IVXVDefineAcknowledgement{}
\global\let\@IVXVAcknowledgement\@empty

% Pagestyle with headers and footers
\setlength{\headwidth}{\textwidth}
\makepagestyle{IVXVFancy}
\makerunningwidth{IVXVFancy}{\headwidth}

% No indent of paragraphs
\setlength{\parindent}{0pt}
\setlength{\parskip}{1.5ex}

% TODO: Single command for this?
\makeoddhead{IVXVFancy}{}{}{
	\IVXVArial 	\tiny \@IVXVStringConfidential}
\makeevenhead{IVXVFancy}{}{}{
	\IVXVArial \tiny \@IVXVStringConfidential}

\makeoddfoot{IVXVFancy}{
	\IVXVArial
	{\tiny
		{\bfseries \hspace{-0.75em}\@title} \\
			\@date}}
	{}
	{
	\IVXVArial
	{\tiny
		\@IVXVDocumentVersion \\
		\thepage\ / \pageref{LastPage}}}
\makeevenfoot{IVXVFancy}{
	\IVXVArial
	{\tiny
		{\bfseries \hspace{-0.75em}\@title} \\
			\@date}}
	{}
	{
	\IVXVArial
	{\tiny
		\@IVXVDocumentVersion \\
		\thepage\ / \pageref{LastPage}}}

\makefootrule{IVXVFancy}{\headwidth}{\normalrulethickness}{0pt}

% We have multiline headers, therefore less space is available
% for text.
\addtolength{\footskip}{2mm}

% Enforce pagestyle for pages with new chapters as well
\aliaspagestyle{chapter}{IVXVFancy}

% Set the new style as the default
\pagestyle{IVXVFancy}

% Titlepage style with headings
\makepagestyle{IVXVTitle}
\makerunningwidth{IVXVTitle}{\headwidth}
\makeoddhead{IVXVTitle}{
	\begin{minipage}[t]{0.5\textwidth}
		\hspace{-7mm}
	\end{minipage}
	}{}{
	\raisebox{2.5mm}{\IVXVArial \scriptsize \bfseries \@IVXVStringConfidential}
	}
% Chapter and section headline style

\renewcommand{\printchapternum}{\LARGE \bfseries \IVXVArial \thechapter}
\renewcommand{\printchaptertitle}[1]{\LARGE \bfseries \IVXVArial #1}
\setlength{\beforechapskip}{60pt}
\setlength{\afterchapskip}{40pt}

\let\stdchapter\chapter
\renewcommand\chapter{\clearpage\ \stdchapter}

\setsecheadstyle{\normalsize \bfseries \IVXVArial}
\setsubsecheadstyle{\normalsize \bfseries \IVXVArial}
\setsubsubsecheadstyle{\IVXVArial}

\usepackage{graphicx}
% title page
\makeatletter
\renewcommand{\maketitle}{%
	\thispagestyle{IVXVTitle}
	% \vspace{2cm}

	\begin{minipage}[t]{0.5\textwidth}
		%\vspace{-6.5mm}
		%\hfill
		\vspace{31mm}
	\end{minipage}
	\vspace{4cm}
	\begin{flushleft}
		\IVXVArial
		{\LARGE \bfseries \@title}
		\vfill
		{\Large \bfseries
			\@IVXVStringDocumentType

			\ifx\@IVXVDocumentVersion\@empty
				Dok. versiooninumber defineerimata. Kasuta IVXVDefineDocumentVersion käsku või IVXV-Document-Version stiili.
			\else
				\@IVXVStringVersion~\@IVXVDocumentVersion
			\fi

			\@date

			\pageref{LastPage} \@IVXVStringPages\\[1.5ex]

			\ifx\@IVXVDocumentNumber\@empty
				Dok. hoidla ning järjenumber defineerimata. Kasuta IVXVDefineDocumentNumber käsku või IVXV-Document-Number stiili.
			\else
				\@IVXVStringDocument~\@IVXVDocumentNumber
			\fi
		}\par
		\vspace{2cm}
		{\scriptsize \@IVXVAcknowledgement}
	\end{flushleft}
	\pagebreak
%	\thispagestyle{empty}
%	\@IVXVStringBackofTitlePage
%	\cleardoublepage
}
\makeatother

% Make dots appear in the table of contents also for chapters
\renewcommand{\cftchapterdotsep}{\cftdotsep}

% URL formating, URL breaking and linked entries in the PDF

\usepackage{url}
%%% XXX remove breakurl
% \usepackage{breakurl}
%%% XXX \usepackage[pdfborder={0 0 0}, pdftex, pdfpagelayout=TwoPageRight]{hyperref}
% hyperref and memoir is incompatible, this fixes the problems
\usepackage{memhfixc}

% make enumerate and itemize lists tighter
\firmlists
% and change the numbering for nested lists
\renewcommand{\labelenumi}{\arabic{enumi}.}
\renewcommand{\labelenumii}{\arabic{enumi}.\arabic{enumii}}
\renewcommand{\labelenumiii}{\arabic{enumi}.\arabic{enumii}.\arabic{enumiii}}
\renewcommand{\labelenumiv}{\arabic{enumi}.\arabic{enumii}.\arabic{enumiii}.\arabic{enumiii}.\arabic{enumiv}}
% make LaTeX observe right margins at all costs:
\sloppy
