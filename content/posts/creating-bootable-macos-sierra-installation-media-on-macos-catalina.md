---
title: "Creating bootable macOS Sierra installation media on macOS Catalina"
date: 2020-08-24T23:34:00-07:00
draft: false
---

If, like a friend of mine, you find yourself needing to install a fresh copy of
macOS Sierra on an older Mac, and you want to create a bootable USB flash drive
to hold the installer using `createinstallmedia`, a quick bit of Googling starts
to turn up numerous threads [like this](https://discussions.apple.com/thread/250811969).

**tl;dr:** In late 2019, Apple published an updated macOS Sierra installer which
broke the ability to create bootable installation media, failing with an error
like `/Volumes/XYZ is not a valid volume mount point`.

Attempting to use Sierra's `createinstallmedia` on my macOS Catalina machine
yielded a slightly different set of problems, but with a bit of
reverse-engineering, I was able to figure out how to make it work.

If you don't care for the nitty gritty technical analysis, you can [skip right to
the instructions below](#just-give-me-the-instructions-dammit).

## technical deep-dive

The oft-suggested workaround, assuming you have an older Mac upon which the
Sierra installer can be installed, involves finding and downloading an older
copy of `Install macOS Sierra.app` from a who-knows-how-trustworthy source,
running `createinstallmedia` from within the newer installer app bundle, and
then renaming the older installer app bundle out from under `createinstallmedia`
before confirming that you'd like to erase the target disk.

Gross.

With those instructions, if you _don't_ have an older Mac upon which you can
install the Sierra installer (or if it's not in working order), you're kind of
out of luck.

I neither had an older Mac, nor was I content with downloading a
possibly-sketchy older version of the installer from some random place on the
internet, so I set my sights on figuring out how to prepare the installation
medium from my Catalina machine.

### problem #1: missing `InstallESD.dmg`

After downloading `InstallOS.dmg` via the _Download macOS Sierra_ link in step 4
of [this Apple support document][support-doc], I mounted the disk image, and
attempted to run the installation package, only to be met with this dialog:

![Error dialog: This version of macOS 10.12 cannot be installed on this computer.](/images/Screen-Shot-2020-08-24-at-6.10.56-PM.png)

Oops. My machine (a 2019 Mac mini) is too new to run Sierra, so the installer
package is preventing me from even installing `Install macOS Sierra.app`.

OK, fine—we'll extract the contents of the installer package manually:

```
% pkgutil --expand-full /Volumes/Install\ macOS/InstallOS.pkg ~/Desktop/InstallOS
```

Note that we use pkgutil's undocumented `--expand-full` flag here, because
`--expand` isn't enough to cause `Install macOS Sierra.app` to be extracted from
the `Payload` file within the installer package.

Now let's copy the installer app to `/Applications`:

```
% sudo ditto ~/Desktop/InstallOS/InstallOS.pkg/Payload/Install\ macOS\ Sierra.app \
    /Applications/Install\ macOS\ Sierra.app
```

And attempt to run `createinstallmedia` (assuming we have a blank flash drive
mounted at `/Volumes/Untitled`):

```
% sudo /Applications/Install\ macOS\ Sierra.app/Contents/Resources/createinstallmedia \
    --volume /Volumes/Untitled \
    --applicationpath /Applications/Install\ macOS\ Sierra.app
```

Immediately, we get `/Applications/Install macOS Sierra.app does not appear to be
a valid OS installer application`. Grr.

Using [Hopper][hopper] to disassemble `createinstallmedia`, I quickly narrowed
in on the cause of the `does not appear to be a valid OS installer application`
error message, centered around a failing validation routine called at address
`0x1000017a3`:

```
00000001000017a0         mov        rdi, r13
00000001000017a3         call       sub_100001cb2
00000001000017a8         cmp        eax, 0x1
00000001000017ab         je         loc_1000018b7
00000001000017b1         test       eax, eax
00000001000017b3         je         loc_10000198c
```

If the validation routine at `sub_100001cb2` returns `1`, we jump to
`0x1000018b7` and continue on. If it returns `0`, we jump to `0x10000198c`, at
which point the `does not appear to be a valid OS installer application` message
is printed, and the program exits. If it returns anything other than `0` or `1`,
we don't jump anywhere, and execute the next instruction after `0x1000017b3`.

`sub_100001cb2`, as Hopper calls it, isn't too terribly difficult to read, but
rather than regurgitating a wall of x86\_64 assembly here, I'll just explain
what it does. It:

- Accepts an `NSString *` as its only parameter, which is the path to `Install
  macOS Sierra.app` that the user specified for the `--applicationpath` flag to
  `createinstallmedia`
- Validates the existence of that path, and that it's a directory
- Ensures that the bundle identifier of the application begins with
  `com.apple.InstallAssistant`
- Looks for `SharedSupport/OSInstall.mpkg` within the app bundle, and if it
  can't find that, falls back to looking for `SharedSupport/InstallInfo.plist`
- If `InstallInfo.plist` exists, it then looks up the `System Image Info -> URL`
  key, and looks for a disk image named after that value, which is expected to
  contain `OSInstall.mpkg`

Here's our first problem.

`InstallInfo.plist` looks like this:

```
% plutil -p /Applications/Install\ macOS\ Sierra.app/Contents/SharedSupport/InstallInfo.plist
{
  "Additional Installers" => [
  ]
  "Additional Wrapped Installers" => [
  ]
  "OS Installer" => "OSInstall.mpkg"
  "System Image Info" => {
    "id" => "com.apple.dmg.InstallESD"
    "sha1" => ""
    "URL" => "InstallESD.dmg"
    "version" => "10.12.6"
  }
}
```

`System Image Info -> URL` tells us we're looking for `InstallESD.dmg`, and
`sub_100001cb2` expects that path to be relative to the `SharedSupport`
directory.  Let's see what's in `SharedSupport`:

```
% ls /Applications/Install\ macOS\ Sierra.app/Contents/SharedSupport 
InstallInfo.plist
```

D'oh! No `InstallESD.dmg`.

After a bit of head scratching, I realized that, because `InstallOS.pkg` never
ran successfully, its `link_package` script never ran, which contains this very
important line:

```sh
/bin/cp "${PACKAGE_PATH}" "${SHARED_SUPPORT_PATH}/InstallESD.dmg"
```

Easy fix—since we've already extracted `InstallOS.pkg`, we can copy
`InstallESD.dmg` over manually:

```
cp ~/Desktop/InstallOS/InstallOS.pkg/InstallESD.dmg \
    /Applications/Install\ macOS\ Sierra.app/Contents/SharedSupport
```

Cool, that's that out of the way.

### problem #2: mismatched bundle version strings

Now, at this point, you might be tempted to run `createinstallmedia` again.

Don't do that.

Without fixing one last issue, attempting to run `createinstallmedia` will cause
it to spawn copies of itself ad infinitum until your machine runs out of RAM, or
you `sudo killall -9` it.

That's because of one additonal check in our friend `sub_100001cb2`. The
pseudocode looks like this:

```objc
loc_100001dc1:
    r14 = 0x0;
    rdx = rax;
    if ([r13 fileExistsAtPath:rdx] != 0x0) {
            rbx = [[r12 infoDictionary] objectForKey:@"CFBundleShortVersionString"];
            r14 = 0x1;
            if ([var_38 isEqualToString:@"com.apple.InstallAssistant.Sierra"] != 0x0) {
                    CMP([rbx isEqualToString:@"12.6.03"], 0x1);
                    r14 = 0x1 - 0xffffffff + CARRY(RFLAGS(cf));
            }
    }
    rax = r14;
    return rax;
```

`r13` is an `NSFileManager *`, rdx is `/Applications/Install macOS
Sierra.app/Contents/SharedSupport/InstallESD.dmg`, and `var_38` is the bundle
identifier of `Install macOS Sierra.app` (`com.apple.InstallAssistant.Sierra`).

If `InstallESD.dmg` exists (which it should, now that we've copied it into the
app bundle), and if the bundle identifier is `com.apple.InstallAssistant.Sierra`
(it is), then we check to see if the `CFBundleShortVersionString` is `12.6.03`.
Depending on the result of that comparison, `r14` is set, and is then returned
in `rax`.

The Hopper-generated pseudocode for the assignment to `r14` looks a little
weird—the assembly looks like this:

```
0000000100001e39         cmp        al, 0x1
0000000100001e3b         mov        r14d, 0x1
0000000100001e41         sbb        r14d, 0xffffffff
```

My x86\_64 assembly-fu was just rusty enough that that didn't make immediate
sense to me, and I didn't feel like launching `createinstallmedia` under `lldb`
to debug interactively, so I whipped up a quick test program to help me
understand what that code was doing:

```objc
int main(int argc, const char * argv[])
{
    @autoreleasepool {
        BOOL result = [@"12.6.03" isEqualToString:@"12.6.03"];
        int32_t var = 0;
        asm (
            "cmpb $0x1, %1\n"
            "movl $0x1, %0\n"
            "sbbl $0xffffffff, %0\n"
            : "=r" (var)
            : "r" (result)
        );
        printf("result: %d; var: %d\n", result, var);
    }
    return 0;
}
```

That gives us:

```
result: 1; var: 2
```

Changing either of the strings in the `-isEqualToString:` comparison so that
they _don't_ match gives us:

```
result: 0; var: 1
```

Interesting—so our validation routine at `sub_100001cb2` must be returning `1`
or `2` (the former if the strings don't match, and the latter if they do match).
Recall that if `sub_100001cb2` returns `0`, we get the `/Applications/Install
macOS Sierra.app does not appear to be a valid OS installer application` error.

So let's see if our `Info.plist` has the bundle version we're expecting. Survey
says:

```
% plutil -p /Applications/Install\ macOS\ Sierra.app/Contents/Info.plist \
    | grep CFBundleShortVersionString
  "CFBundleShortVersionString" => "12.6.06"
```

Aha! So `sub_100001cb2` must be returning `1`, which results in us jumping to
`loc_1000018b7`, and doing this:

```objc
loc_1000018b7:
    rax = [NSBundle bundleWithPath:r13];
    rax = [rax pathForResource:@"createinstallmedia" ofType:0x0];
    var_40 = rax;
    if (rax != 0x0) {
            r14 = var_38 - 0x1;
            var_30 = [NSMutableArray arrayWithCapacity:sign_extend_64(r14)];
            if (var_38 >= 0x2) {
                    rbx = rbx + 0x8;
                    do {
                            [var_30 addObject:[NSString stringWithUTF8String:*rbx]];
                            rbx = rbx + 0x8;
                            r14 = r14 - 0x1;
                    } while (r14 != 0x0);
            }
            r14 = [NSTask launchedTaskWithLaunchPath:var_40 arguments:var_30];
            [r14 waitUntilExit];
            exit([r14 terminationStatus]);
    }
    else {
            fprintf(**___stderrp, "%s does not appear to be a valid OS installer application.\n", [r13 UTF8String]);
            exit(0xfffffffffffffffd);
    }
    return;
```

Meaning, this copy of `createinstallmedia` will spawn a copy of itself as a child
process with identical arguments, and wait for the child to exit. Since it's
been given the same arguments as its parent, the child process executes in
exactly the same way, and ends up reaching this code, where it too spawns a copy
of itself, and waits for _that_ process to exit. And on and on... 😐

Thankfully, this is also an easy fix: we can just modify `Install macOS
Sierra.app/Contents/Info.plist` to have the “correct” bundle version string:

```
% sed -i '' -e 's/12\.6\.06/12.6.03/' \
    /Applications/Install\ macOS\ Sierra.app/Contents/Info.plist
```

Verify that did the trick:

```
% plutil -p /Applications/Install\ macOS\ Sierra.app/Contents/Info.plist \
    | grep CFBundleShortVersionString
  "CFBundleShortVersionString" => "12.6.03"
```

Brilliant.

Moment of truth—let's try running `createinstallmedia` again:

```
% sudo /Applications/Install\ macOS\ Sierra.app/Contents/Resources/createinstallmedia \
    --volume /Volumes/Untitled \
    --applicationpath /Applications/Install\ macOS\ Sierra.app
Ready to start.
To continue we need to erase the disk at /Volumes/Untitled.
If you wish to continue type (Y) then press return:
```

Yahtzee! That seems to have done the trick.

## just give me the instructions, dammit

Note that I performed these steps on a machine running macOS Catalina (10.15.6).
I don't see any reason why these steps wouldn't work on older versions of macOS,
but, y'know, YMMV.

- Download the macOS Sierra `InstallOS.dmg` from step 4 of [this Apple support
  document][support-doc] by clicking the _Download macOS Sierra_ link
- Mount the newly-downloaded `InstallOS.dmg` disk image by double-clicking it
- Extract the contents of the installer package to your desktop:

```
pkgutil --expand-full /Volumes/Install\ macOS/InstallOS.pkg \
    ~/Desktop/InstallOS
```

- Copy `Install macOS Sierra.app` to your `/Applications` folder:

```
sudo ditto ~/Desktop/InstallOS/InstallOS.pkg/Payload/Install\ macOS\ Sierra.app \
    /Applications/Install\ macOS\ Sierra.app
```

- Copy `InstallESD.dmg` to the `SharedSupport` directory within the installer
  app bundle:

```
cp ~/Desktop/InstallOS/InstallOS.pkg/InstallESD.dmg \
    /Applications/Install\ macOS\ Sierra.app/Contents/SharedSupport
```

- Fix the bundle version string in `Install macOS Sierra.app`'s `Info.plist`:

```
sed -i '' -e 's/12\.6\.06/12.6.03/' \
    /Applications/Install\ macOS\ Sierra.app/Contents/Info.plist
```

- Finally, run `createinstallmedia`:

```
sudo /Applications/Install\ macOS\ Sierra.app/Contents/Resources/createinstallmedia \
    --volume /Volumes/Untitled \
    --applicationpath /Applications/Install\ macOS\ Sierra.app
```

The last step assumes you've got your target flash drive mounted at
`/Volumes/Untitled`. If necessary, replace that path with wherever your flash
drive is mounted. `createinstallmedia` will format and overwrite that drive, so
Type Carefully™.

Look at you! You've got the macOS Sierra installer on a bootable flash drive.

[support-doc]: https://support.apple.com/en-us/HT208202
[hopper]: https://www.hopperapp.com
