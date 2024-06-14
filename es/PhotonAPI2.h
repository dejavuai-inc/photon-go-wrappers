// Photon API For Go / C++
// Copyright (c) DejaVuAI, 2000-2024
// All rights reserved
#include <stdio.h>
#include <stdlib.h>

#pragma once

// Search modes
#define SEARCH_MODE_GRAY 0
#define SEARCH_MODE_RGB 1  // not supported yet
#define SEARCH_MODE_BOTH 2 // not supported yet

// Rotation
#define SEARCH_ROTATION_NONE 0
#define SEARCH_ROTATION_MUL_90 1
#define SEARCH_ROTATION_ARB 2

// Sensitivity
#define SEARCH_SENSITIVITY_NORM 0
#define SEARCH_SENSITIVITY_HIGH 1

// Search options/flags
#define SEARCH_MIRRORED 1
#define SEARCH_THOROUGH 2
#define SEARCH_GET_IMAGE_INFO 4

// Search staging enumerator
#define SEARCH_SINGLE_STAGE 0
#define SEARCH_DOUBLE_STAGE_BASIC 1
#define SEARCH_DOUBLE_STAGE_OPTIMIZED 2
#define SEARCH_DOUBLE_STAGE_DEFAULT SEARCH_DOUBLE_STAGE_BASIC

// Resolution levels
#define MIN_RESOLUTION_LEVEL 2
#define MAX_RESOLUTION_LEVEL 12
#define DEFAULT_RESOLUTION_LEVEL 4

#define SERVER_HEARTBEAT_LOST 1

extern void statusCallback(int status, void *param);

#ifdef __cplusplus
extern "C"
{
#endif
    struct PhotonAPI2_BufInfo
    {
        unsigned char *data;
        unsigned int dataSize;
    };
    struct PhotonAPI2_StrInfo
    {
        char *str;
        unsigned int length;
    };

    int Init();

    int Uninit();

    void Disconnect(void *ctx);

    int GetInfo(void *ctx, struct PhotonAPI2_StrInfo *pStrInfo);

    int Search(void *ctx, const char *fileName, const unsigned char *imageData, unsigned int imageDataSize, unsigned int resolution, unsigned short rotation, unsigned int sensitivity, unsigned int flags, struct PhotonAPI2_StrInfo *pStrInfo);

    void FreeString(struct PhotonAPI2_StrInfo *pStrInfo);

    typedef void (*PFPhotonAPI2_StatusCallback)(int status, void *param);

    void *Connect(const char *address, PFPhotonAPI2_StatusCallback statusCB, void *statusCBParam);

    int GetThumbnail(void *ctx, const char *sessionGUID, unsigned int imageID, const char *encoding, struct PhotonAPI2_StrInfo *pStrInfo, struct PhotonAPI2_BufInfo *pBufInfo);

    void FreeBuffer(struct PhotonAPI2_BufInfo *pBufInfo);

#ifdef __cplusplus
}
#endif
