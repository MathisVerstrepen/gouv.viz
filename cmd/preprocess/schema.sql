CREATE TABLE dataset_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE acteurs (
  uid TEXT PRIMARY KEY,
  civilite TEXT,
  prenom TEXT,
  nom TEXT,
  alpha TEXT,
  date_naissance TEXT,
  ville_naissance TEXT,
  dep_naissance TEXT,
  pays_naissance TEXT,
  date_deces TEXT,
  profession TEXT,
  uri_hatvp TEXT,
  source_file TEXT,
  source_hash TEXT
);

CREATE TABLE acteur_adresses (
  uid TEXT PRIMARY KEY,
  acteur_uid TEXT NOT NULL REFERENCES acteurs(uid),
  type_code TEXT,
  type_libelle TEXT,
  poids INTEGER,
  adresse_rattachement TEXT,
  valeur TEXT
);

CREATE TABLE organes (
  uid TEXT PRIMARY KEY,
  code_type TEXT,
  libelle TEXT,
  libelle_abrege TEXT,
  libelle_abrev TEXT,
  libelle_edition TEXT,
  legislature INTEGER,
  chambre TEXT,
  regime TEXT,
  organe_parent_uid TEXT REFERENCES organes(uid),
  position_politique TEXT,
  couleur_associee TEXT,
  preseance INTEGER,
  date_debut TEXT,
  date_agrement TEXT,
  date_fin TEXT,
  secretaire_01_uid TEXT,
  secretaire_02_uid TEXT,
  source_file TEXT,
  source_hash TEXT
);

CREATE TABLE mandats (
  uid TEXT PRIMARY KEY,
  acteur_uid TEXT NOT NULL REFERENCES acteurs(uid),
  legislature INTEGER,
  type_organe TEXT,
  date_debut TEXT,
  date_fin TEXT,
  date_publication TEXT,
  preseance INTEGER,
  nomin_principale INTEGER,
  code_qualite TEXT,
  lib_qualite TEXT,
  lib_qualite_sex TEXT
);

CREATE TABLE mandat_organes (
  mandat_uid TEXT NOT NULL REFERENCES mandats(uid),
  organe_uid TEXT NOT NULL REFERENCES organes(uid),
  PRIMARY KEY (mandat_uid, organe_uid)
);

CREATE TABLE scrutins (
  uid TEXT PRIMARY KEY,
  numero INTEGER NOT NULL,
  legislature INTEGER NOT NULL,
  organe_uid TEXT REFERENCES organes(uid),
  session_ref TEXT,
  seance_ref TEXT,
  date_scrutin TEXT NOT NULL,
  quantieme_jour_seance INTEGER,
  code_type_vote TEXT,
  libelle_type_vote TEXT,
  type_majorite TEXT,
  sort_code TEXT,
  sort_libelle TEXT,
  titre TEXT,
  demandeur_texte TEXT,
  objet_libelle TEXT,
  mode_publication_votes TEXT,
  nombre_votants INTEGER,
  suffrages_exprimes INTEGER,
  suffrages_requis INTEGER,
  non_votants INTEGER,
  pour INTEGER,
  contre INTEGER,
  abstentions INTEGER,
  non_votants_volontaires INTEGER,
  source_file TEXT,
  source_hash TEXT
);

CREATE TABLE scrutin_groupe_votes (
  scrutin_uid TEXT NOT NULL REFERENCES scrutins(uid),
  groupe_uid TEXT NOT NULL REFERENCES organes(uid),
  nombre_membres_groupe INTEGER,
  position_majoritaire TEXT,
  non_votants INTEGER,
  pour INTEGER,
  contre INTEGER,
  abstentions INTEGER,
  non_votants_volontaires INTEGER,
  PRIMARY KEY (scrutin_uid, groupe_uid)
);

CREATE TABLE votes (
  id INTEGER PRIMARY KEY,
  scrutin_uid TEXT NOT NULL REFERENCES scrutins(uid),
  acteur_uid TEXT NOT NULL REFERENCES acteurs(uid),
  mandat_uid TEXT REFERENCES mandats(uid),
  groupe_uid TEXT REFERENCES organes(uid),
  position TEXT NOT NULL,
  par_delegation INTEGER,
  num_place TEXT,
  UNIQUE (scrutin_uid, acteur_uid)
);

CREATE TABLE acteur_vote_stats (
  acteur_uid TEXT NOT NULL REFERENCES acteurs(uid),
  legislature INTEGER NOT NULL,
  total_votes INTEGER NOT NULL,
  pour INTEGER NOT NULL,
  contre INTEGER NOT NULL,
  abstentions INTEGER NOT NULL,
  non_votants INTEGER NOT NULL,
  PRIMARY KEY (acteur_uid, legislature)
);

CREATE TABLE groupe_vote_stats (
  groupe_uid TEXT NOT NULL REFERENCES organes(uid),
  legislature INTEGER NOT NULL,
  total_scrutins INTEGER NOT NULL,
  pour INTEGER NOT NULL,
  contre INTEGER NOT NULL,
  abstentions INTEGER NOT NULL,
  non_votants INTEGER NOT NULL,
  PRIMARY KEY (groupe_uid, legislature)
);

CREATE VIRTUAL TABLE scrutin_search USING fts5(
  uid UNINDEXED,
  document,
  tokenize = 'unicode61 remove_diacritics 2'
);

CREATE INDEX idx_scrutins_date ON scrutins(date_scrutin);
CREATE INDEX idx_scrutins_legislature ON scrutins(legislature);
CREATE INDEX idx_scrutins_sort ON scrutins(sort_code);
CREATE INDEX idx_votes_scrutin ON votes(scrutin_uid);
CREATE INDEX idx_votes_acteur ON votes(acteur_uid);
CREATE INDEX idx_votes_groupe ON votes(groupe_uid);
CREATE INDEX idx_votes_position ON votes(position);
CREATE INDEX idx_mandats_acteur ON mandats(acteur_uid);
CREATE INDEX idx_mandats_dates ON mandats(date_debut, date_fin);
CREATE INDEX idx_mandat_organes_organe ON mandat_organes(organe_uid);
CREATE INDEX idx_organes_code_type ON organes(code_type);
CREATE INDEX idx_organes_legislature ON organes(legislature);
