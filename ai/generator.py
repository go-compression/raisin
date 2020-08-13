import random

from PIL import Image, ImageDraw
from fpdf import FPDF


def generate_image(name, text, width=128, height=128):
    rand_pixels = [random.randint(0, 255) for _ in range(width * height * 3)]
    rand_pixels_as_bytes = bytes(rand_pixels)

    random_image = Image.frombytes('RGB', (width, height), rand_pixels_as_bytes)

    draw_image = ImageDraw.Draw(random_image)
    draw_image.text(xy=(0, 0), text=text, fill=(255, 255, 255))
    random_image.save("{file_name}.jpg".format(file_name=name))
    return name + ".jpg"


def generate_pdf(name, text):
    pdf = FPDF()
    pdf.add_page()
    pdf.set_font('Arial', 'B', 16)
    pdf.cell(40, 10, text)
    pdf.output(name + '.pdf', 'F')
    return name + '.pdf'