#version 330 core
in vec3 aPos;
in vec2 aTexCoord;

out vec2 texCoord;

uniform mat4 projection;

void main()
{    
    gl_Position = projection * vec4(aPos, 1.0);
    texCoord = aTexCoord;
}
