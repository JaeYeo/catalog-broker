package com.infranics.cloudvm;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;

@Slf4j
@Component
public class ShellCommandExecutor {
	/**
	 * exec new sub-process
	 */
	public void exec(Path dir_path, String command) throws Exception {
		ProcessBuilder builder = new ProcessBuilder();

		builder.directory(dir_path.toFile());

		// check os
		if (isWindow()) {
			builder.command("cmd.exe", "/c", command);
		} else {
			builder.command("sh", "-c", command);
		}

		// set error stream
		builder.redirectErrorStream(true);

		// start process
		Process process = builder.start();

		// get input stream
		InputStream is = process.getInputStream();

		BufferedReader br = new BufferedReader(new InputStreamReader(is, StandardCharsets.UTF_8.name()));
		String line = br.readLine();

		while (line != null) {
			// log
			log.info(line);

			// read line
			line = br.readLine();
		}

		// waiting
		process.waitFor();

		// check exit code
		int exitCode = process.exitValue();

		if (exitCode == 0) {
			//@ success
		} else if (exitCode == 126) {
			// permission denied
			throw new Exception("permission denied");
		} else if (exitCode == 127) {
			// file not found
			throw new Exception(command + " not found");
		} else {
			// unknown
			throw new Exception("unknown exception[" + exitCode + "]" );
		}
	}

	/**
	 * is window?
	 */
	private boolean isWindow() {
		return System.getProperty("os.name").toLowerCase().startsWith("windows");
	}
}
