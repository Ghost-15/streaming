# Schémas UML & BPMN (Ce3.6.1)

Représentation de l'architecture, des modèles de données et des processus métier de
StreamPulse — plateforme de streaming audio live développée en Flutter (mobile et web),
Go (backend) et Supabase (PostgreSQL).

## 1. Diagramme de classes — Modèles Flutter

Source : `flutter/lib/api/models/`

```mermaid
classDiagram
    direction LR
    class UserModel {
        +String id
        +String email
        +String firstName
        +String lastName
        +Role role
        +DateTime suspendedAt
        +bool isAdmin()
        +bool isDiffuseur()
        +bool isSuspended()
    }
    class StreamModel {
        +String id
        +String title
        +String broadcasterId
        +String broadcasterName
        +int listenerCount
        +String streamUrl
        +String activeSessionId
        +bool isLive
        +DateTime createdAt
        +copyWith()
    }
    class PlaylistModel {
        +String id
        +String ownerId
        +String title
        +bool isQueue
        +int trackCount
        +DateTime createdAt
        +List~TrackModel~ tracks
    }
    class TrackModel {
        +String id
        +String title
        +String artist
        +int duration
        +String fileUrl
        +int position
        +DateTime createdAt
    }
    class Role {
        <<enumeration>>
        anon
        user
        diffuseur
        admin
    }
    UserModel --> Role : a
    UserModel "1" --> "*" StreamModel : diffuse
    UserModel "1" --> "*" PlaylistModel : possède
    PlaylistModel "1" o-- "*" TrackModel : contient
```

Les modèles Flutter correspondent aux entités retournées par l'API Go. `TrackModel`
est utilisé pour les items de playlist et de favoris, qui référencent des streams
depuis la migration 011.

## 2. Diagramme de composants — Architecture globale

```mermaid
graph TB
    subgraph Flutter["Application Flutter (Mobile + Web)"]
        direction TB
        UI["Screens & Widgets\n(home, library, broadcaster, admin)"]
        NOT["Notifiers — ChangeNotifier + Provider\n(StreamNotifier, AudioNotifier, PlaylistNotifier…)"]
        REPO["Repositories\n(StreamRepository, PlaylistRepository…)"]
        SVC["ApiService — HTTP Client\nStorageService — SharedPreferences"]
    end
    subgraph Backend["Go Backend (Clean Architecture)"]
        direction TB
        HDL["Handlers HTTP\n(auth, stream, playlist, favorite, admin)"]
        UC["Use Cases\n(logique métier, recommandations)"]
        REP["Repositories Go\n(SQL via pgx)"]
    end
    subgraph Infra["Supabase"]
        DB[("PostgreSQL + RLS\n11 migrations")]
        STG["Storage Bucket\n(audio WebM/Opus)"]
        AUTH["JWT Auth"]
    end
    UI --> NOT
    NOT --> REPO
    REPO --> SVC
    SVC -->|"REST/JSON HTTPS"| HDL
    HDL --> UC
    UC --> REP
    REP --> DB
    REP --> STG
    DB -.-> AUTH
```

Le backend Go adopte une architecture clean plate : Handlers → Use Cases → Repositories →
Entités. L'application Flutter suit le même principe : Widgets → Notifiers (Provider) →
Repositories → ApiService.

## 3. Diagramme de séquence — Cycle de vie d'un stream

```mermaid
sequenceDiagram
    actor D as Diffuseur
    participant F as App Flutter
    participant G as Go Backend
    participant DB as Supabase DB
    actor A as Auditeur

    D->>F: Démarre la diffusion
    F->>G: POST /streams (titre)
    G->>DB: INSERT INTO streams (status='live')
    DB-->>G: stream_id + active_session_id
    G-->>F: StreamModel (id, sessionId, streamUrl)

    loop Diffusion en cours
        D->>F: Audio capturé (WebM/Opus)
        F->>G: POST /streams/{id}/chunks
        G->>DB: Stockage chunk (sessionId vérifié)
    end

    A->>F: Parcourt la liste des streams
    F->>G: GET /streams?status=live
    G-->>F: []StreamModel
    A->>F: Sélectionne un stream
    F->>G: POST /listen_history (event=join)
    G->>DB: INSERT listen_history
    G-->>F: URL audio
    F->>F: Lecture via media_kit / Web Audio API

    D->>F: Arrête la diffusion
    F->>G: PATCH /streams/{id} (status=ended)
    G->>DB: UPDATE streams SET status='ended', active_session_id=NULL
```

`active_session_id` est un UUID renouvelé à chaque diffusion : il empêche les chunks
d'une session précédente de corrompre la diffusion en cours (migration 010).

## 4. BPMN — Parcours utilisateur : écouter un stream

```mermaid
flowchart TD
    Start([Ouverture de l'app]) --> Login[Écran de connexion]
    Login -->|Identifiants invalides| ErrLogin[Afficher erreur]
    ErrLogin --> Login
    Login -->|Authentifié avec succès| Home[Accueil — liste des streams live]
    Home --> Available{Stream disponible ?}
    Available -->|Non| Refresh[Actualiser / patienter]
    Refresh --> Available
    Available -->|Oui| Select[Sélectionner un stream]
    Select --> Join[Rejoindre le stream\nEvent join enregistré]
    Join --> Listen[Écoute active\nMiniPlayer visible]
    Listen --> UserAction{Action utilisateur ?}
    UserAction -->|Ajouter aux favoris| Fav[Enregistrer en favori]
    Fav --> Listen
    UserAction -->|Ajouter à une playlist| AddPl[Choisir une playlist]
    AddPl --> Listen
    UserAction -->|Quitter| Leave[Quitter le stream\nEvent leave enregistré]
    Leave --> ReturnHome[Retour à l'accueil]
    ReturnHome --> End([Session terminée])
```

Les événements `join` et `leave` sont enregistrés dans `listen_history` pour alimenter
le moteur de recommandations (US-025).
