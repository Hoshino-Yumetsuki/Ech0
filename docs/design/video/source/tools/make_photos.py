# Original procedural "photos" for demo Echos (no third-party imagery). Deterministic seeds.
import numpy as np
from PIL import Image, ImageFilter
rng = np.random.default_rng(7)
def grain(a, s=6):
    return np.clip(a + rng.normal(0, s, a.shape), 0, 255)
def save(a, name, blur=0):
    im = Image.fromarray(grain(a).astype('uint8'))
    if blur: im = im.filter(ImageFilter.GaussianBlur(blur))
    im.save(f'public/demo/{name}.jpg', quality=90)

def lerp(c1, c2, t): return c1[None, None, :] * (1 - t[..., None]) + c2[None, None, :] * t[..., None]

# 1. harbor at dawn: warm sky, sun disk, glittering water
W, H = 1600, 1066
y, x = np.mgrid[0:H, 0:W] / np.array([H, W])[:, None, None]
hz = 0.56
sky = lerp(np.array([242, 203, 160.]), np.array([250, 236, 214.]), np.clip(y / hz, 0, 1) ** 0.8)
sky = lerp(np.array([196, 150, 128.]), np.array([250, 226, 190.]), np.clip(y / hz, 0, 1) ** 0.6) * 0.5 + sky * 0.5
d = np.hypot((x - 0.62) * 1.5, (y - hz + 0.06))
sky += (np.exp(-d * 18) * 90)[..., None] * np.array([1, .85, .6])
sea = lerp(np.array([120, 118, 124.]), np.array([62, 74, 92.]), np.clip((y - hz) / (1 - hz), 0, 1))
glint = (rng.random((H, W)) > 0.985) * np.exp(-np.abs(x - 0.62) * 9) * np.clip((y - hz) * 6, 0, 1)
glint = np.array(Image.fromarray((glint * 255).astype('uint8')).filter(ImageFilter.GaussianBlur(1.2))) / 255.
sea += (glint * 900)[..., None] * np.array([1, .9, .75])
img = np.where((y < hz)[..., None], sky, sea)
save(img, 'harbor')

# 2. dusk ridge: layered hills in haze
W, H = 1600, 1200
y, x = np.mgrid[0:H, 0:W] / np.array([H, W])[:, None, None]
img = lerp(np.array([236, 196, 170.]), np.array([246, 230, 210.]), y)
cols = [np.array(c, float) for c in ([196, 150, 140], [150, 110, 108], [104, 80, 82], [60, 50, 56])]
for i, c in enumerate(cols):
    base = 0.42 + i * 0.13
    ridge = base + 0.05 * np.sin(x * (6 + i * 3) + i * 1.7) + 0.025 * np.sin(x * (19 + i * 5) + i)
    img = np.where((y > ridge)[..., None], c[None, None, :] * np.ones_like(img), img)
save(img, 'ridge', blur=0.6)

# 3. desk at night: lamp glow on a dark wooden surface with a notebook
W, H = 1600, 1200
y, x = np.mgrid[0:H, 0:W] / np.array([H, W])[:, None, None]
img = lerp(np.array([40, 30, 26.]), np.array([86, 58, 40.]), np.clip(y * 1.2, 0, 1))
img += (np.sin(x * 90 + np.sin(y * 7) * 3) * 6)[..., None]
glow = np.exp(-np.hypot(x - 0.3, y - 0.2) * 3.2)
img += (glow * 170)[..., None] * np.array([1, .78, .5])
nb = (np.abs(x - 0.58) < 0.2) & (np.abs(y - 0.6) < 0.22)
img = np.where(nb[..., None], np.array([236, 226, 208.]) * (0.75 + glow[..., None] * 0.6), img)
lines = nb & (np.mod(y * 60, 1) < 0.06) & (np.abs(x - 0.58) < 0.17)
img = np.where(lines[..., None], img * 0.78, img)
save(img, 'desk', blur=0.8)
print('ok')
