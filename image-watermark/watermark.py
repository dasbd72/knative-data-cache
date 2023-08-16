from PIL import Image, ImageDraw, ImageFont

def add_watermark(input_image_path, output_image_path, watermark_text, font_size=40, opacity=100):
    try:
        # Open the input image
        image = Image.open(input_image_path)

        # Create a drawing object
        draw = ImageDraw.Draw(image)

        # Choose a font and size for the watermark
        font = ImageFont.truetype("arial.ttf", font_size)

        # Get the size of the image
        image_width, image_height = image.size

        # Get the size of the watermark text
        text_width, text_height = draw.textsize("Water Mark", font)

        # Calculate the position to place the watermark (centered)
        x = (image_width - text_width) // 2
        y = (image_height - text_height) // 2

        # Add the watermark to the image
        draw.text((x, y), watermark_text, font=font, fill=(255, 255, 255, opacity))

        # Save the watermarked image
        image.save(output_image_path)

        print("Watermark added successfully.")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    input_image_path = "path/to/input/image.jpg"
    output_image_path = "path/to/output/watermarked_image.jpg"

    add_watermark(input_image_path, output_image_path)
