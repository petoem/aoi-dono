# Aoi-dono

Post from your favourite text `$EDITOR` to Bluesky and Mastodon. Simply run `aoi-dono` in your shell, write your post, save it, close and you are done! 

Just like `git commit` ✨

## How to get `Aoi`?

```sh
go install github.com/petoem/aoi-dono@latest
```

Alternatively, you can download a binary release [here](https://github.com/petoem/aoi-dono/releases).

## How to setup and use `Aoi`?

### Configure your favourite text editor ...

... by setting one of the following shell environment variables `AOI_EDITOR`, `VISUAL`, `EDITOR` to your editors executable.

### Set your auth credentials and settings ...

... by commandline flags or environment variables. You can use flag `-save-to-config` to save your current credentials and settings to a config file in `XDG_CONFIG_HOME`, this way you don't have to pass them in on the commandline.

See `aoi-dono -help` for available options.

### How to get your credentials?

#### Bluesky

You need an app password for the account you want to post to. In your Bluesky account settings, navigate to _Privacy and Security_ > _App passwords_.

#### Mastodon

You will need a token for Mastodon. In your account preferences, navigate to _Development_ > _New Application_. Give your new token the `write:statuses` scope.

## Contributing

If you wish to contribute to the code or documentation, feel free to fork the repository and submit a pull request.

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
