# Gbege

This is a backend service for a mobile application that allows users to covertly upload audio files to a server.

## What is Gbege?

Gbege is a mobile application that allows users to covertly upload audio files to a server. The app is designed to be used in situations where users may be in danger and need to record evidence of abuse or other crimes.

## How does it work?

- A user creates a profile and creates spaces where uploads will be sent.
- There will be feeds where users can comment or upload media to.
- There will be public and private spaces.
- Private spaces are only available to permitted users.
- Public spaces are available to all users.
- Recordings are done in the background without any visibility for a certain duration(20mins).
- Recordings are activated by voice i.e use of preconfigured words(2).
- Phone vibrates after upload with no visible ui.
- After secretly uploading, the user has time(30mins) to vet before it is available for listeners.
- There will be ability to download, delete(owner only) and share(owner only).
- Uploads will be deleted after 6 months.
- There should be ability to hide the app icon (show with voice).

## Tech Stack

- Go
- Postgres
- Redis
- React Native (Expo)
- Containerized using Podman
