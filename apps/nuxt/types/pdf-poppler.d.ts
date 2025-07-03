declare module "pdf-poppler" {
  interface ConvertOptions {
    format: "png" | "jpeg" | "tiff" | "pdf";
    out_dir: string;
    out_prefix: string;
    page?: number;
    scale?: number;
    quality?: number;
  }

  export function convert(
    pdfPath: string,
    options: ConvertOptions
  ): Promise<void>;
  export function info(pdfPath: string): Promise<any>;
  export function imgdata(pdfPath: string): Promise<any>;
  export const path: string;
  export const exec_options: {
    encoding: string;
    maxBuffer: number;
    shell: boolean;
  };
}
