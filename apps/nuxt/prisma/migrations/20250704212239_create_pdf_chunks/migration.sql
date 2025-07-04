/*
  Warnings:

  - You are about to drop the `PDF` table. If the table is not empty, all the data it contains will be lost.

*/
-- DropTable
DROP TABLE "PDF";

-- CreateTable
CREATE TABLE "pdfs" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "url" TEXT NOT NULL,
    "coverUrl" TEXT,
    "totalPages" INTEGER NOT NULL,
    "isArchived" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "pdfs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "pdf_chunks" (
    "id" TEXT NOT NULL,
    "pdfId" TEXT NOT NULL,
    "index" INTEGER NOT NULL,
    "text" TEXT NOT NULL,
    "audioUrl" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "pdf_chunks_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "pdf_chunks_pdfId_idx" ON "pdf_chunks"("pdfId");

-- CreateIndex
CREATE UNIQUE INDEX "pdf_chunks_pdfId_index_key" ON "pdf_chunks"("pdfId", "index");

-- AddForeignKey
ALTER TABLE "pdf_chunks" ADD CONSTRAINT "pdf_chunks_pdfId_fkey" FOREIGN KEY ("pdfId") REFERENCES "pdfs"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
