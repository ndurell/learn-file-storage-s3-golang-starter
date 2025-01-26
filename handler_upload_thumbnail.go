package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, _, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	defer file.Close()

	imageData, err := io.ReadAll(file)

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content Type", nil)
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Video ID", err)
	}

	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User not authorized for video", nil)
	}

	videoThumbnails[videoID] = thumbnail{
		data:      imageData,
		mediaType: contentType,
	}

	videoURL := fmt.Sprintf("http://localhost:8091/api/thumbnails/%s", videoID)
	videoData.ThumbnailURL = &videoURL
	err = cfg.db.UpdateVideo(videoData)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving video", err)

	}
	respondWithJSON(w, http.StatusOK, struct{}{})
}
