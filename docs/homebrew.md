# Homebrew Installation Guide

This document explains how to set up a Homebrew tap for ekslogs.

## Creating a Homebrew Tap Repository

1. Create a new GitHub repository named `homebrew-ekslogs`
   - The repository name must start with `homebrew-`
   - Make it public

2. Clone the repository locally:
   ```bash
   git clone https://github.com/kzcat/homebrew-ekslogs.git
   cd homebrew-ekslogs
   ```

3. Create a Formula file named `ekslogs.rb` with the following content:
   ```ruby
   class Ekslogs < Formula
     desc "A fast and intuitive CLI tool for retrieving and monitoring Amazon EKS cluster Control Plane logs"
     homepage "https://github.com/kzcat/ekslogs"
     url "https://github.com/kzcat/ekslogs/archive/refs/tags/v0.1.4.tar.gz"
     sha256 "REPLACE_WITH_ACTUAL_SHA256"
     license "MIT"
     
     depends_on "go" => :build
     
     def install
       system "go", "build", *std_go_args(ldflags: "-s -w")
     end
     
     test do
       assert_match "ekslogs version", shell_output("#{bin}/ekslogs version")
     end
   end
   ```

4. Calculate the SHA256 checksum of the tarball:
   ```bash
   curl -L https://github.com/kzcat/ekslogs/archive/refs/tags/v0.1.4.tar.gz | shasum -a 256
   ```

5. Replace `REPLACE_WITH_ACTUAL_SHA256` with the actual SHA256 checksum.

6. Commit and push the Formula:
   ```bash
   git add ekslogs.rb
   git commit -m "Add ekslogs formula"
   git push
   ```

7. Test the Formula locally:
   ```bash
   brew tap kzcat/ekslogs
   brew install ekslogs
   ```

## Updating the Formula

When a new version of ekslogs is released:

1. Update the URL and SHA256 in the Formula:
   ```ruby
   url "https://github.com/kzcat/ekslogs/archive/refs/tags/vX.Y.Z.tar.gz"
   sha256 "NEW_SHA256_CHECKSUM"
   ```

2. Commit and push the changes:
   ```bash
   git add ekslogs.rb
   git commit -m "Update ekslogs to vX.Y.Z"
   git push
   ```

## Automating Formula Updates

Consider setting up a GitHub Action in the main ekslogs repository that automatically updates the Formula in the homebrew-ekslogs repository when a new release is created.
