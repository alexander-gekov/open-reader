-- CreateTable
CREATE TABLE "PDF" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "description" TEXT,
    "pageCount" INTEGER NOT NULL,
    "fileSize" INTEGER NOT NULL,
    "fileKey" TEXT NOT NULL,
    "coverKey" TEXT,
    "chunks" TEXT[],
    "audioId" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "lastReadPage" INTEGER NOT NULL DEFAULT 0,
    "lastReadAt" TIMESTAMP(3),
    "isArchived" BOOLEAN NOT NULL DEFAULT false,
    "metadata" JSONB,

    CONSTRAINT "PDF_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "PDF_userId_idx" ON "PDF"("userId");

-- CreateIndex
CREATE INDEX "PDF_fileKey_idx" ON "PDF"("fileKey");
