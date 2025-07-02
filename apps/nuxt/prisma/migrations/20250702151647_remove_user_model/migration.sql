-- CreateTable
CREATE TABLE "TTSSettings" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "provider" TEXT NOT NULL DEFAULT 'elevenlabs',
    "apiKey" TEXT NOT NULL,
    "model" TEXT,
    "voice" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "TTSSettings_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "TTSSettings_userId_key" ON "TTSSettings"("userId");

-- CreateIndex
CREATE INDEX "TTSSettings_userId_idx" ON "TTSSettings"("userId");
