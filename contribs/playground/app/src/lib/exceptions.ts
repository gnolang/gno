export class InvalidFileNameException extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'InvalidFileNameException'
  }
}

export class DuplicateFileNameException extends InvalidFileNameException {
  constructor() {
    super('File name already exists. Please choose a different name.')
  }
}

export class InvalidNameException extends InvalidFileNameException {
  constructor() {
    super(`File name is invalid. Only letters, numbers, and the following characters are allowed: -_()`)
  }
}

export class InvalidExtensionException extends Error {
  constructor() {
    super('Invalid file extension. Only .gno and .toml extensions are supported')
  }
}

export class ReservedFileNameException extends InvalidFileNameException {
  constructor() {
    super('This file name is reserved. Please choose a different name.')
  }
}
