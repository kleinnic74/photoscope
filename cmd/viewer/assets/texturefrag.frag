#version 420 core

in vec2 texCoord;
out vec4 fragColor;

uniform sampler2D t;

void main() {
    fragColor = texture(t, texCoord);
}

