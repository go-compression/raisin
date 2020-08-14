import random
import uuid

from essential_generators import DocumentGenerator
from fpdf import FPDF
from PIL import Image, ImageDraw


def generate_files(training_set):
    files = []
    for training_items in training_set:
        generator = training_items[0]
        quantity = training_items[1]
        param_generator = training_items[2]

        for i in range(quantity):
            args, kwargs = param_generator()
            file = generator(*args, **kwargs)

            files.append(file)
    return files

def generate_jpg(name, text, width=128, height=128):
    rand_pixels = [random.randint(0, 255) for _ in range(width * height * 3)]
    rand_pixels_as_bytes = bytes(rand_pixels)

    random_image = Image.frombytes('RGB', (width, height), rand_pixels_as_bytes)

    draw_image = ImageDraw.Draw(random_image)
    draw_image.text(xy=(0, 0), text=text, fill=(255, 255, 255))
    random_image.save("{file_name}.jpg".format(file_name=name))
    return name + ".jpg"

def generate_jpg_params():
    name = str(uuid.uuid4())
    return (name, name * 4), {"width": 256, "height": 256}

def generate_pdf(name, text):
    pdf = FPDF()
    pdf.add_page()
    pdf.set_font('Arial', 'B', 16)
    pdf.cell(40, 10, text)
    pdf.output(name + '.pdf', 'F')
    return name + '.pdf'

gen = None

def generate_pdf_params():
    if not gen:
        gen = DocumentGenerator()
    randnum = random.randint(0, 100)
    name = str(uuid.uuid4())
    text = ""
    if randnum < 25:
        text += ",".join([gen.email() for i in range(random.randrange(100, 500))])
        name += "email"
    elif randnum < 50:
        text += ",".join([gen.phone() for i in range(random.randrange(100, 500))])
        name += "phone"
    elif randnum < 75:
        text += ",".join([gen.url() for i in range(random.randrange(100, 500))])
        name += "url"
    else:
        text += ",".join([gen.sentence() for i in range(random.randrange(100, 500))])
        name += "words"
    return (name, ''.join([i if ord(i) < 128 else ' ' for i in text])), {}