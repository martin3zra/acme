const DELIMITERS = ['\t', ',', ';', '|', ':'];

export function useDelimiterDetector() {
  function detectDelimiter(line: string) {
    let best = ',';
    let maxCount = 0;

    for (const d of DELIMITERS) {
      const count = line.split(d).length - 1;
      if (count > maxCount) {
        maxCount = count;
        best = d;
      }
    }

    return best;
  }

  async function detectFromFile(file: File) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();

      reader.onload = (e) => {
        // Read only the first line — enough to detect
        const firstLine = (e.target?.result as string).split(/\r?\n/)[0];
        resolve(detectDelimiter(firstLine));
      };

      reader.onerror = reject;

      // Read just the first 1KB — no need to load the whole file
      reader.readAsText(file.slice(0, 1024));
    });
  }

  return { detectFromFile };
}
