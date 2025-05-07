package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

const maxMemory int64 = 10 << 20

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
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}

	thumbnailFile, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find thumbnail", err)
		return
	}
	headerContentType := header.Header.Get("Content-Type")
	if headerContentType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}

	// b, err := io.ReadAll(thumbnailFile)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Couldn't read thumbnail", err)
	// 	return
	// }

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to find the video", err)
		return 
	}
	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	
	mediaType, _, err := mime.ParseMediaType(headerContentType)
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "incorrect file type", fmt.Errorf("incorrect file type"))
		return
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse media type", err)
		return
	}


	fileType := strings.Split(mediaType, "/")
	fileString := videoID.String() + "." 
	if len(fileType) > 1 {
		fileString += fileType[1]
	} else {
		respondWithError(w, http.StatusBadRequest, "invalid file type", fmt.Errorf("invalid file type"))
		return
	}
	
	filePath := filepath.Join(cfg.assetsRoot, fileString)

	file, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to upload file", err)
		return
	}
	defer file.Close()

	io.Copy(file, thumbnailFile)

	dataURL := fmt.Sprintf("http://localhost:%s/%s", cfg.port, filePath)

	videoData.ThumbnailURL = &dataURL

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to update the video data", err)
		return 
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
