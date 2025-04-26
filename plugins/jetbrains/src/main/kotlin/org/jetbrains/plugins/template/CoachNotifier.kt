package org.jetbrains.plugins.template

import com.intellij.openapi.editor.event.*
import com.intellij.openapi.fileEditor.FileEditorManagerListener
import com.intellij.openapi.util.Key
import kotlinx.coroutines.*
import org.json.JSONObject
import java.net.URI
import java.net.http.HttpClient
import java.net.http.WebSocket
import java.util.concurrent.CompletionStage

private val KEY_CF_THOUGHTS: Key<String> = Key.create("coach.thoughts")

private val PROBLEM_REGEX =
    Regex("""//\s*problem:\s*([0-9]+[A-Z][0-9]?)""", RegexOption.IGNORE_CASE)

// Matches any single-line comment that starts with // and captures the comment text that follows.
private val COMMENT_REGEX = Regex("""//\s?(.*)""")

class CoachNotifier : FileEditorManagerListener {
    private var ws: WebSocket? = null
    private val scope = CoroutineScope(Dispatchers.IO)

    override fun fileOpened(manager: com.intellij.openapi.fileEditor.FileEditorManager, file: com.intellij.openapi.vfs.VirtualFile) {
        val editor = manager.getSelectedTextEditor() ?: return
        val debounceMs = 2000L
        var job: Job? = null

        editor.document.addDocumentListener(object : DocumentListener {
            override fun documentChanged(event: DocumentEvent) {

                job?.cancel()
                job = scope.launch {
                    delay(debounceMs)

                    val text = editor.document.text

                    // Extract the problem identifier if present
                    val problemMatch = PROBLEM_REGEX.find(text)
                    val problemId = problemMatch?.groupValues?.getOrNull(1)?.uppercase()

                    // Collect every line comment as a "thought"
                    val thoughts = COMMENT_REGEX.findAll(text)
                        .map { it.groupValues[1].trim() }
                        .filter { it.isNotEmpty() }
                        .joinToString("\n")

                    // Store thoughts in the editor's user data so other components can access them if needed
                    editor.putUserData(KEY_CF_THOUGHTS, thoughts)

                    // Do not proceed if there is no problem ID detected
                    if (problemId == null) return@launch

                    // Build JSON payload
                    val msg = JSONObject()
                        .put("problemId", problemId)
                        .put("code", text)
                        .put("thoughts", thoughts)

                    // Lazy WebSocket connection
                    if (ws == null) {
                        ws = HttpClient.newHttpClient()
                            .newWebSocketBuilder()
                            .buildAsync(
                                URI("ws://localhost:12345/ws"),
                                object : WebSocket.Listener {
                                    override fun onText(
                                        webSocket: WebSocket?, data: CharSequence?, last: Boolean
                                    ): CompletionStage<*>? = null
                                }
                            ).join()
                    }
                    ws?.sendText(msg.toString(), true)
                }
            }
        })
    }
}
