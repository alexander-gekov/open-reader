/*
  Warnings:

  - You are about to drop the column `audioId` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `chunks` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `coverKey` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `description` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `fileKey` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `fileSize` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `lastReadAt` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `lastReadPage` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `metadata` on the `PDF` table. All the data in the column will be lost.
  - You are about to drop the column `pageCount` on the `PDF` table. All the data in the column will be lost.
  - Added the required column `totalPages` to the `PDF` table without a default value. This is not possible if the table is not empty.
  - Added the required column `url` to the `PDF` table without a default value. This is not possible if the table is not empty.

*/
-- DropIndex
DROP INDEX "PDF_fileKey_idx";

-- AlterTable
ALTER TABLE "PDF" ADD COLUMN "url" TEXT;
ALTER TABLE "PDF" ADD COLUMN "coverUrl" TEXT;
ALTER TABLE "PDF" ADD COLUMN "totalPages" INTEGER;

-- Update existing records with default values
UPDATE "PDF" SET 
  "url" = CONCAT('https://storage.example.com/', "fileKey"),
  "totalPages" = "pageCount";

-- Make the columns required after setting default values
ALTER TABLE "PDF" ALTER COLUMN "url" SET NOT NULL;
ALTER TABLE "PDF" ALTER COLUMN "totalPages" SET NOT NULL;

-- Drop old columns
ALTER TABLE "PDF" DROP COLUMN "fileKey",
                  DROP COLUMN "fileSize",
                  DROP COLUMN "pageCount",
                  DROP COLUMN "lastReadPage",
                  DROP COLUMN "lastReadAt",
                  DROP COLUMN "audioId",
                  DROP COLUMN "chunks",
                  DROP COLUMN "coverKey",
                  DROP COLUMN "description",
                  DROP COLUMN "metadata";
