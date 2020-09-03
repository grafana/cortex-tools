// build_image is the image used by default to build make targets.
local build_image = 'golang:1.14.8-stretch';

// make defines the common configuration for a Drone step that builds a make target.
local make(target) = {
  name: 'make-%s' % target,
  image: build_image,
  commands: ['make %s' % target],
};

// image_tag defines a step that runs the image-tag script and writes it to a file that can be used by subsequent docker steps. See http://plugins.drone.io/drone-plugins/drone-docker/ for how the `.tags` file works.
local image_tag() = {
  name: 'image-tag',
  image: 'alpine',
  commands: [
    'apk add --no-cache bash git',
    'git fetch origin --tags',
    'echo $(./tools/image-tag)  > .tags',
  ],
};

// pipeline defines an empty Drone pipeline.
local pipeline(name) = {
  kind: 'pipeline',
  name: name,
  steps: [
    image_tag(),
  ],
};

[
  // Run 
  pipeline('validate-pull-request') {
    steps+: [
      make('lint'),
      make('test'),
      make('all'),
    ],
    trigger: {
      event: [
        "pull_request",
      ],
    },
  },
]
