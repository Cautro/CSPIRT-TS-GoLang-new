export async function compressImage(
    file: File,
    maxWidth = 512,
    maxHeight = 512,
    quality = 0.75
): Promise<string> {
    const imageBitmap = await createImageBitmap(file);

    let { width, height } = imageBitmap;

    const ratio = Math.min(maxWidth / width, maxHeight / height, 1);
    width = Math.round(width * ratio);
    height = Math.round(height * ratio);

    const canvas = document.createElement("canvas");
    canvas.width = width;
    canvas.height = height;

    const ctx = canvas.getContext("2d");
    if (!ctx) {
        throw new Error("Canvas context is not available");
    }

    ctx.drawImage(imageBitmap, 0, 0, width, height);

    return canvas.toDataURL("image/webp", quality);
}