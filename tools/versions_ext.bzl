"""Module extension that parses versions.env into a Starlark dict.

This makes versions.env the single source of truth for all tool versions.
Bazel consumers load from @versions//:versions.bzl instead of a hand-maintained file.
"""

def _versions_repo_impl(rctx):
    env_content = rctx.read(rctx.attr.versions_file)
    versions = {}
    for line in env_content.split("\n"):
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split("=", 1)
        if len(parts) != 2:
            continue
        key = parts[0].strip()
        if key.endswith("_VERSION"):
            key = key[:-len("_VERSION")]
        versions[key] = parts[1].strip()

    content = '"""Auto-generated from versions.env — do not edit."""\n\nVERSIONS = {\n'
    for k in sorted(versions.keys()):
        content += '    "{}": "{}",\n'.format(k, versions[k])
    content += "}\n"

    rctx.file("versions.bzl", content)
    rctx.file("BUILD.bazel", "")

_versions_repo = repository_rule(
    implementation = _versions_repo_impl,
    attrs = {
        "versions_file": attr.label(allow_single_file = True),
    },
)

def _versions_ext_impl(module_ctx):
    _versions_repo(
        name = "versions",
        versions_file = "@//:versions.env",
    )

versions_ext = module_extension(implementation = _versions_ext_impl)
