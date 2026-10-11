import { useEffect, useState } from "react";

export type SongRecord = {
  id?: number;
  title?: string;
  artist?: string;
  artists?: string[] | string | { String?: string; Valid?: boolean };
  album?: string | { String?: string; Valid?: boolean };
  image_url?: string | { String?: string; Valid?: boolean };
  yt_url?: string | { String?: string; Valid?: boolean };
  spotify_url?: string | { String?: string; Valid?: boolean };
  duration_ms?: number | { Int32?: number; Valid?: boolean };
  created_at?: string;
};

export type Match = SongRecord;

interface MatchResultProps {
  match: SongRecord;
  onReset: () => void;
}

const safeStr = (val: unknown): string => {
  if (!val) return "";
  if (typeof val === "string") return val;
  if (typeof val === "object" && val !== null && "String" in val) {
    const obj = val as { String?: string; Valid?: boolean };
    return obj.Valid !== false ? obj.String || "" : "";
  }
  return String(val);
};

const safeArtists = (val: unknown): string => {
  if (!val) return "";
  if (Array.isArray(val)) {
    return val.map((v) => safeStr(v)).filter(Boolean).join("، ");
  }
  if (typeof val === "string") {
    try {
      const parsed = JSON.parse(val);
      if (Array.isArray(parsed)) return parsed.map((v) => safeStr(v)).join("، ");
    } catch {
      return val;
    }
    return val;
  }
  return safeStr(val);
};

const MatchResult = ({ match, onReset }: MatchResultProps) => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  const title = safeStr(match?.title);
  const artists = safeArtists(match?.artists);
  const album = safeStr(match?.album);
  const imageUrl = safeStr(match?.image_url);
  const ytUrl =
    safeStr(match?.yt_url) ||
    `https://www.youtube.com/results?search_query=${encodeURIComponent(
      `${title} ${artists}`.trim()
    )}`;
  const spotifyUrl =
    safeStr(match?.spotify_url) ||
    `https://open.spotify.com/search/${encodeURIComponent(
      `${title} ${artists}`.trim()
    )}`;

  return (
    <div
      dir="rtl"
      className={`overflow-hidden absolute inset-0 z-50 flex flex-col items-center justify-between bg-black/90 text-white p-6 transition-opacity duration-1000 ${
        isVisible ? "opacity-100" : "opacity-0"
      }`}
    >
      {imageUrl ? (
        <div
          className="absolute inset-0 bg-cover bg-center opacity-40 blur-3xl scale-125 z-0"
          style={{ backgroundImage: `url(${imageUrl})` }}
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-b from-blue-900/40 via-black to-black z-0" />
      )}
      <div className="absolute inset-0 bg-gradient-to-b from-black/30 via-transparent to-black/90 z-0" />

      <div className="relative z-10 w-full flex justify-between items-center pt-4">
        <button
          onClick={onReset}
          aria-label="واپس وڃو"
          className="p-2 rounded-full bg-white/10 backdrop-blur-md hover:bg-white/20 transition-all cursor-pointer"
        >
          <svg
            className="w-6 h-6 text-white"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </button>
        <span className="text-sm font-medium text-white/70">گيت سڃاڻپ نتيجو</span>
        <div className="w-10" />
      </div>

      <div className="relative z-10 flex flex-col items-center w-full max-w-sm space-y-6 my-auto">
        <div
          className={`relative w-64 h-64 rounded-2xl overflow-hidden shadow-2xl shadow-blue-500/20 transition-all duration-1000 delay-300 bg-white/5 flex items-center justify-center ${
            isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
          }`}
        >
          {imageUrl ? (
            <img
              src={imageUrl}
              alt={album || title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="flex flex-col items-center justify-center text-blue-400">
              <svg className="w-24 h-24" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z" />
              </svg>
            </div>
          )}
        </div>

        <div
          className={`text-center space-y-2 transition-all duration-1000 delay-500 ${
            isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
          }`}
        >
          <h1 className="text-3xl font-bold leading-tight line-clamp-2">
            {title || "اڻڄاتل گيت"}
          </h1>
          {artists && (
            <p className="text-xl text-blue-400 font-medium">
              {artists}
            </p>
          )}
          {album && (
            <p className="text-sm text-gray-400 uppercase tracking-wider mt-1">
              {album}
            </p>
          )}
        </div>
      </div>

      <div
        className={`relative z-10 w-full max-w-sm space-y-2.5 pb-8 transition-all duration-1000 delay-700 ${
          isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
        }`}
      >
        {/* Media Action Buttons */}
        <div className="grid grid-cols-2 gap-2 w-full">
          <a
            href={ytUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="py-3 bg-red-600 hover:bg-red-700 text-white font-bold rounded-xl transition-all flex items-center justify-center gap-2 shadow-lg text-xs"
          >
            <svg className="w-4 h-4 fill-current shrink-0" viewBox="0 0 24 24">
              <path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/>
            </svg>
            یوٽيوب (YouTube)
          </a>

          <a
            href={spotifyUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="py-3 bg-emerald-600 hover:bg-emerald-700 text-white font-bold rounded-xl transition-all flex items-center justify-center gap-2 shadow-lg text-xs"
          >
            <svg className="w-4 h-4 fill-current shrink-0" viewBox="0 0 24 24">
              <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.498 17.306c-.215.352-.676.463-1.028.247-2.816-1.721-6.36-2.111-10.535-1.157-.403.092-.807-.16-.899-.562-.092-.403.16-.807.562-.899 4.568-1.044 8.5-1.602 11.652.343.353.216.464.676.248 1.028zm1.468-3.262c-.27.44-.847.578-1.287.308-3.224-1.982-8.138-2.555-11.95-1.398-.496.15-1.023-.136-1.173-.632-.15-.496.136-1.023.632-1.173 4.356-1.322 9.776-.682 13.47 1.608.44.27.578.847.308 1.287zm.126-3.41c-3.863-2.294-10.237-2.505-13.914-1.389-.592.18-1.22-.164-1.4-.756-.18-.592.164-1.22.756-1.4 4.227-1.284 11.264-1.037 15.696 1.593.533.316.708 1.01.392 1.543-.316.533-1.01.708-1.53.409z"/>
            </svg>
            اسپوٽيفائي (Spotify)
          </a>
        </div>

        <a
          href={`https://www.google.com/search?q=${encodeURIComponent(
            `${title} ${artists} سنڌي گيت`.trim()
          )}`}
          target="_blank"
          rel="noopener noreferrer"
          className="w-full py-2.5 bg-white/15 hover:bg-white/25 backdrop-blur-md text-white font-medium rounded-xl transition-all flex items-center justify-center gap-2 border border-white/15 shadow-sm text-xs"
        >
          <svg className="w-4 h-4 fill-current shrink-0" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/>
          </svg>
          گوگل تي ڳوليو (Search on Web)
        </a>

        <button
          onClick={onReset}
          className="w-full py-3 bg-white/10 hover:bg-white/20 backdrop-blur-md text-white font-semibold rounded-full transition-transform active:scale-95 cursor-pointer border border-white/15 flex items-center justify-center gap-2"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          ٻيو گيت سڃاڻو
        </button>
      </div>
    </div>
  );
};

export default MatchResult;
