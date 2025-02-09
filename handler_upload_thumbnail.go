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
	if err = r.ParseMultipartForm(maxMemory); err!=nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse multipart form", err)
		return
	}

	file, handler, err := r.FormFile("thumbnail")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Couldn't find thumbnail file", err)
		return
	}

	defer file.Close()

	mediaType:= handler.Header.Get("Content-Type")

	imageData, err:= io.ReadAll(file)
	if err!=nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't read thumbnail", err)
        return
	}

	metaData, err:= cfg.db.GetVideo(videoID)
	if err!=nil{
        respondWithError(w, http.StatusInternalServerError, "Couldn't get video metadata", err)
        return
    }

	if userID != metaData.UserID{
		respondWithError(w, http.StatusUnauthorized, "You can't upload a thumbnail for this video", err)
        return
	}

	videoThumbnails[videoID] = thumbnail{
		data: imageData,
		mediaType: mediaType,
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID)
	metaData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(metaData)
	if err!=nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video metadata", err)
        return
	}

	respondWithJSON(w, http.StatusOK, metaData)

}
